package sofia

import (
   "bytes"
   "encoding/binary"
   "errors"
)

// --- MOOV ---
type MoovChild struct {
   Mvhd *MvhdBox
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

type MoovBox struct {
   Header   BoxHeader
   Children []MoovChild
}

// IsAudio checks the handler type within the first track to determine if it's audio.
func (b *MoovBox) IsAudio() bool {
   if trak, ok := b.Trak(); ok {
      if mdia, ok := trak.Mdia(); ok {
         for _, child := range mdia.Children {
            // Check if the raw box is an 'hdlr' box.
            // The handler_type is at offset 16 of the box content.
            if len(child.Raw) >= 20 && string(child.Raw[4:8]) == "hdlr" {
               handlerType := string(child.Raw[16:20])
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
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child MoovChild
      switch string(header.Type[:]) {
      case "mvhd":
         var mvhd MvhdBox
         if err := mvhd.Parse(content); err != nil {
            return err
         }
         child.Mvhd = &mvhd
      case "trak":
         var trak TrakBox
         if err := trak.Parse(content); err != nil {
            return err
         }
         child.Trak = &trak
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MoovBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mvhd != nil {
         buffer = append(buffer, child.Mvhd.Encode()...)
      } else if child.Trak != nil {
         buffer = append(buffer, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         // Skipped
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *MoovBox) RemovePssh() {
   var kept []MoovChild
   for _, child := range b.Children {
      if child.Pssh != nil {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) RemoveMvex() {
   var kept []MoovChild
   for _, child := range b.Children {
      if len(child.Raw) >= 8 && string(child.Raw[4:8]) == "mvex" {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) Trak() (*TrakBox, bool) {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak, true
      }
   }
   return nil, false
}

func (b *MoovBox) Mvhd() (*MvhdBox, bool) {
   for _, child := range b.Children {
      if child.Mvhd != nil {
         return child.Mvhd, true
      }
   }
   return nil, false
}

func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, child := range b.Children {
      if child.Pssh != nil && bytes.Equal(child.Pssh.SystemID[:], systemID) {
         return child.Pssh, true
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
   b.Version = data[8]
   copy(b.Flags[:], data[9:12])
   offset := 12
   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mvhd v1 too short")
      }
      b.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   } else { // Version 0
      if len(data) < 32 {
         return errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }
   if offset < int(b.Header.Size) {
      b.RemainingData = make([]byte, int(b.Header.Size)-offset)
      copy(b.RemainingData, data[offset:b.Header.Size])
   }
   return nil
}

func (b *MvhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MvhdBox) Encode() []byte {
   var baseSize uint32
   if b.Version == 1 {
      baseSize = 44
   } else {
      baseSize = 32
   }
   totalSize := baseSize + uint32(len(b.RemainingData))
   buffer := make([]byte, totalSize)
   binary.BigEndian.PutUint32(buffer[0:4], totalSize)
   copy(buffer[4:8], b.Header.Type[:])
   buffer[8] = b.Version
   copy(buffer[9:12], b.Flags[:])
   offset := 12
   if b.Version == 1 {
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.CreationTime)
      offset += 8
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.ModificationTime)
      offset += 8
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint64(buffer[offset:offset+8], b.Duration)
      offset += 8
   } else {
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.CreationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.ModificationTime))
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint32(buffer[offset:offset+4], uint32(b.Duration))
      offset += 4
   }
   copy(buffer[offset:], b.RemainingData)
   b.Header.Size = totalSize
   return buffer
}
