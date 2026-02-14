package sofia

import "errors"

// --- TRAK ---
type TrakBox struct {
   Header      BoxHeader
   Mdia        *MdiaBox
   RawChildren [][]byte
}

func (b *TrakBox) Parse(data []byte) error {
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
      case "mdia":
         var mdia MdiaBox
         if err := mdia.Parse(content); err != nil {
            return err
         }
         b.Mdia = &mdia
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

func (b *TrakBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Mdia != nil {
      buffer = append(buffer, b.Mdia.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *TrakBox) RemoveEdts() {
   var kept [][]byte
   for _, child := range b.RawChildren {
      if len(child) >= 8 && string(child[4:8]) == "edts" {
         continue
      }
      kept = append(kept, child)
   }
   b.RawChildren = kept
}

// --- MDIA ---
type MdiaBox struct {
   Header      BoxHeader
   Mdhd        *MdhdBox
   Minf        *MinfBox
   RawChildren [][]byte
}

func (b *MdiaBox) Parse(data []byte) error {
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
      case "mdhd":
         var mdhd MdhdBox
         if err := mdhd.Parse(content); err != nil {
            return err
         }
         b.Mdhd = &mdhd
      case "minf":
         var minf MinfBox
         if err := minf.Parse(content); err != nil {
            return err
         }
         b.Minf = &minf
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

func (b *MdiaBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Mdhd != nil {
      buffer = append(buffer, b.Mdhd.Encode()...)
   }
   if b.Minf != nil {
      buffer = append(buffer, b.Minf.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- MDHD ---
type MdhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   Language         [2]byte
   Quality          [2]byte
}

func (b *MdhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("mdhd box too small")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])

   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mdhd v1 too short")
      }
      b.CreationTime = p.Uint64()
      b.ModificationTime = p.Uint64()
      b.Timescale = p.Uint32()
      b.Duration = p.Uint64()
   } else { // Version 0
      if len(data) < 32 {
         return errors.New("mdhd v0 too short")
      }
      b.CreationTime = uint64(p.Uint32())
      b.ModificationTime = uint64(p.Uint32())
      b.Timescale = p.Uint32()
      b.Duration = uint64(p.Uint32())
   }

   if len(data) < p.offset+4 {
      return errors.New("mdhd truncated at language/quality")
   }
   copy(b.Language[:], p.Bytes(2))
   copy(b.Quality[:], p.Bytes(2))
   return nil
}

func (b *MdhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MdhdBox) Encode() []byte {
   var size uint32
   if b.Version == 1 {
      size = 44
   } else {
      size = 32
   }
   buffer := make([]byte, size)
   w := writer{buf: buffer}

   w.PutUint32(size)
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

   w.PutBytes(b.Language[:])
   w.PutBytes(b.Quality[:])

   b.Header.Size = size
   return buffer
}

// --- MINF ---
type MinfBox struct {
   Header      BoxHeader
   Stbl        *StblBox
   RawChildren [][]byte
}

func (b *MinfBox) Parse(data []byte) error {
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
      case "stbl":
         var stbl StblBox
         if err := stbl.Parse(content); err != nil {
            return err
         }
         b.Stbl = &stbl
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

func (b *MinfBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Stbl != nil {
      buffer = append(buffer, b.Stbl.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}
