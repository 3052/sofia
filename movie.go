package sofia

import (
   "bytes"
   "errors"
)

// --- MOOV ---
type MoovBox struct {
   Header      BoxHeader
   Mvhd        *MvhdBox
   Trak        []*TrakBox
   Pssh        []*PsshBox
   RawChildren [][]byte
}

// IsAudio checks the handler type within the first track to determine if it's audio.
func (b *MoovBox) IsAudio() bool {
   if len(b.Trak) > 0 {
      trak := b.Trak[0]
      if trak.Mdia != nil {
         for _, child := range trak.Mdia.RawChildren {
            // Check if the raw box is an 'hdlr' box.
            // The handler_type is at offset 16 of the box content.
            if len(child) >= 20 && string(child[4:8]) == "hdlr" {
               handlerType := string(child[16:20])
               // Handler type for audio is 'soun'
               return handlerType == "soun"
            }
         }
      }
   }
   return false
}

func (b *MoovBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      var header BoxHeader
      if err := header.Parse(payload[offset:]); err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "mvhd":
         var mvhd MvhdBox
         if err := mvhd.Parse(content); err != nil {
            return err
         }
         b.Mvhd = &mvhd
      case "trak":
         var trak TrakBox
         if err := trak.Parse(content); err != nil {
            return err
         }
         b.Trak = append(b.Trak, &trak)
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, &pssh)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

func (b *MoovBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Mvhd != nil {
      buffer = append(buffer, b.Mvhd.Encode()...)
   }
   for _, trak := range b.Trak {
      buffer = append(buffer, trak.Encode()...)
   }
   // pssh is skipped on encode
   for _, raw := range b.RawChildren {
      buffer = append(buffer, raw...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *MoovBox) RemovePssh() {
   b.Pssh = nil
}

func (b *MoovBox) RemoveMvex() {
   var kept [][]byte
   for _, child := range b.RawChildren {
      if len(child) >= 8 && string(child[4:8]) == "mvex" {
         continue
      }
      kept = append(kept, child)
   }
   b.RawChildren = kept
}

func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, pssh := range b.Pssh {
      if bytes.Equal(pssh.SystemID[:], systemID) {
         return pssh, true
      }
   }
   return nil, false
}

// --- MVHD ---
type MvhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   RemainingData    []byte
}

func (b *MvhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("mvhd box too small")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])

   if b.Version == 1 {
      if len(data) < 36 { // 8 header + 4 version/flags + 24 v1 body
         return errors.New("mvhd v1 too short")
      }
      b.CreationTime = p.Uint64()
      b.ModificationTime = p.Uint64()
      b.Timescale = p.Uint32()
      b.Duration = p.Uint64()
   } else { // Version 0
      if len(data) < 24 { // 8 header + 4 version/flags + 12 v0 body
         return errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(p.Uint32())
      b.ModificationTime = uint64(p.Uint32())
      b.Timescale = p.Uint32()
      b.Duration = uint64(p.Uint32())
   }

   b.RemainingData = data[p.offset:b.Header.Size]
   return nil
}

func (b *MvhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MvhdBox) Encode() []byte {
   var bodySize int
   if b.Version == 1 {
      bodySize = 32 // 8+8+4+8 + 4 for ver/flags
   } else {
      bodySize = 20 // 4+4+4+4 + 4 for ver/flags
   }
   totalSize := uint32(8 + bodySize + len(b.RemainingData))
   buffer := make([]byte, totalSize)

   w := writer{buf: buffer}
   w.PutUint32(totalSize)
   w.PutBytes(b.Header.Type[:])
   w.PutByte(b.Version)
   w.PutBytes(b.Flags[:])

   if b.Version == 1 {
      w.PutUint64(b.CreationTime)
      w.PutUint64(b.ModificationTime)
      w.PutUint32(b.Timescale)
      w.PutUint64(b.Duration)
   } else {
      w.PutUint32(uint32(b.CreationTime))
      w.PutUint32(uint32(b.ModificationTime))
      w.PutUint32(b.Timescale)
      w.PutUint32(uint32(b.Duration))
   }

   w.PutBytes(b.RemainingData)
   b.Header.Size = totalSize
   return buffer
}
