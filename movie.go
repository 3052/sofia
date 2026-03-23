// movie.go
package sofia

import (
   "bytes"
   "errors"
)

// --- MOOV ---
type MoovBox struct {
   Header      *BoxHeader
   Mvhd        *MvhdBox
   Trak        []*TrakBox
   Pssh        []*PsshBox
   RawChildren [][]byte
}

func DecodeMoovBox(data []byte) (*MoovBox, error) {
   b := &MoovBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      header, err := DecodeBoxHeader(payload[offset:])
      if err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return nil, errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "mvhd":
         mvhd, err := DecodeMvhdBox(content)
         if err != nil {
            return nil, err
         }
         b.Mvhd = mvhd
      case "trak":
         trak, err := DecodeTrakBox(content)
         if err != nil {
            return nil, err
         }
         b.Trak = append(b.Trak, trak)
      case "pssh":
         pssh, err := DecodePsshBox(content)
         if err != nil {
            return nil, err
         }
         b.Pssh = append(b.Pssh, pssh)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
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
   Header           *BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   RemainingData    []byte
}

func DecodeMvhdBox(data []byte) (*MvhdBox, error) {
   b := &MvhdBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 12 {
      return nil, errors.New("mvhd box too small")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])

   if b.Version == 1 {
      if len(data) < 36 { // 8 header + 4 version/flags + 24 v1 body
         return nil, errors.New("mvhd v1 too short")
      }
      b.CreationTime = p.Uint64()
      b.ModificationTime = p.Uint64()
      b.Timescale = p.Uint32()
      b.Duration = p.Uint64()
   } else { // Version 0
      if len(data) < 24 { // 8 header + 4 version/flags + 12 v0 body
         return nil, errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(p.Uint32())
      b.ModificationTime = uint64(p.Uint32())
      b.Timescale = p.Uint32()
      b.Duration = uint64(p.Uint32())
   }

   b.RemainingData = data[p.offset:b.Header.Size]
   return b, nil
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
