package sofia

import (
   "errors"
   "fmt"
)

// --- ENC (Encrypted Sample Entry) ---
type EncBox struct {
   Header      *BoxHeader
   EntryHeader []byte
   Sinf        *SinfBox
   RawChildren [][]byte
}

func DecodeEncBox(data []byte) (*EncBox, error) {
   b := &EncBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   boxSize := int(b.Header.Size)
   if boxSize == 0 {
      boxSize = len(data)
      b.Header.Size = uint32(boxSize)
   }
   if boxSize < 8 || boxSize > len(data) {
      return nil, fmt.Errorf("enc box size %d invalid for %d bytes of data", boxSize, len(data))
   }

   var entrySize int
   switch string(b.Header.Type[:]) {
   case "enca":
      entrySize = 28
   case "encv":
      entrySize = 78
   default:
      return nil, fmt.Errorf("unknown encryption box type %q", b.Header.Type[:])
   }

   payloadOffset := 8
   if boxSize < payloadOffset+entrySize {
      return nil, fmt.Errorf("enc box too small for sample entry header: need %d bytes, have %d", payloadOffset+entrySize, boxSize)
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+entrySize]

   payload := data[payloadOffset+entrySize : boxSize]
   offset := 0
   for offset < len(payload) {
      header, err := DecodeBoxHeader(payload[offset:])
      if err != nil {
         break
      }
      childSize := int(header.Size)
      if childSize == 0 {
         childSize = len(payload) - offset
      }
      if childSize < 8 || offset+childSize > len(payload) {
         return nil, errors.New("invalid child box size")
      }

      content := payload[offset : offset+childSize]
      switch string(header.Type[:]) {
      case "sinf":
         sinf, err := DecodeSinfBox(content)
         if err != nil {
            return nil, err
         }
         b.Sinf = sinf
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += childSize
   }
   return b, nil
}

func (b *EncBox) Encode() []byte {
   buffer := make([]byte, 8)
   buffer = append(buffer, b.EntryHeader...)
   if b.Sinf != nil {
      buffer = append(buffer, b.Sinf.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- FRMA ---
type FrmaBox struct {
   Header     *BoxHeader
   DataFormat [4]byte
}

func DecodeFrmaBox(data []byte) (*FrmaBox, error) {
   b := &FrmaBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 12 {
      return nil, errors.New("frma box is too small")
   }
   copy(b.DataFormat[:], data[8:12])
   return b, nil
}

func (b *FrmaBox) Encode() []byte {
   buffer := make([]byte, 12)
   w := writer{buf: buffer}
   w.PutUint32(12)
   w.PutBytes([]byte{'f', 'r', 'm', 'a'})
   w.PutBytes(b.DataFormat[:])
   return buffer
}

// --- MDHD ---
type MdhdBox struct {
   Header           *BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   Language         [2]byte
   Quality          [2]byte
}

func DecodeMdhdBox(data []byte) (*MdhdBox, error) {
   b := &MdhdBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 12 {
      return nil, errors.New("mdhd box too small")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])

   if b.Version == 1 {
      if len(data) < 44 {
         return nil, errors.New("mdhd v1 too short")
      }
      b.CreationTime = p.Uint64()
      b.ModificationTime = p.Uint64()
      b.Timescale = p.Uint32()
      b.Duration = p.Uint64()
   } else { // Version 0
      if len(data) < 32 {
         return nil, errors.New("mdhd v0 too short")
      }
      b.CreationTime = uint64(p.Uint32())
      b.ModificationTime = uint64(p.Uint32())
      b.Timescale = p.Uint32()
      b.Duration = uint64(p.Uint32())
   }

   if len(data) < p.offset+4 {
      return nil, errors.New("mdhd truncated at language/quality")
   }
   copy(b.Language[:], p.Bytes(2))
   copy(b.Quality[:], p.Bytes(2))
   return b, nil
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

func (b *MdhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

// --- MDIA ---
type MdiaBox struct {
   Header      *BoxHeader
   Mdhd        *MdhdBox
   Minf        *MinfBox
   RawChildren [][]byte
}

func DecodeMdiaBox(data []byte) (*MdiaBox, error) {
   b := &MdiaBox{}
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
      case "mdhd":
         mdhd, err := DecodeMdhdBox(content)
         if err != nil {
            return nil, err
         }
         b.Mdhd = mdhd
      case "minf":
         minf, err := DecodeMinfBox(content)
         if err != nil {
            return nil, err
         }
         b.Minf = minf
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
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

// --- MINF ---
type MinfBox struct {
   Header      *BoxHeader
   Stbl        *StblBox
   RawChildren [][]byte
}

func DecodeMinfBox(data []byte) (*MinfBox, error) {
   b := &MinfBox{}
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
      case "stbl":
         stbl, err := DecodeStblBox(content)
         if err != nil {
            return nil, err
         }
         b.Stbl = stbl
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
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

// --- SCHI (Scheme Information) ---
type SchiBox struct {
   Header      *BoxHeader
   Tenc        *TencBox
   RawChildren [][]byte
}

func DecodeSchiBox(data []byte) (*SchiBox, error) {
   b := &SchiBox{}
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
      case "tenc":
         tenc, err := DecodeTencBox(content)
         if err != nil {
            return nil, err
         }
         b.Tenc = tenc
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

func (b *SchiBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Tenc != nil {
      buffer = append(buffer, b.Tenc.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- SINF ---
type SinfBox struct {
   Header      *BoxHeader
   Frma        *FrmaBox
   Schi        *SchiBox
   RawChildren [][]byte
}

func DecodeSinfBox(data []byte) (*SinfBox, error) {
   b := &SinfBox{}
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
      case "frma":
         frma, err := DecodeFrmaBox(content)
         if err != nil {
            return nil, err
         }
         b.Frma = frma
      case "schi":
         schi, err := DecodeSchiBox(content)
         if err != nil {
            return nil, err
         }
         b.Schi = schi
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

func (b *SinfBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Frma != nil {
      buffer = append(buffer, b.Frma.Encode()...)
   }
   if b.Schi != nil {
      buffer = append(buffer, b.Schi.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- STSD ---
type StsdBox struct {
   Header       *BoxHeader
   HeaderFields [8]byte // Ver(1)+Flags(3)+EntryCount(4)
   EncChildren  []*EncBox
   RawChildren  [][]byte
}

func DecodeStsdBox(data []byte) (*StsdBox, error) {
   b := &StsdBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 16 {
      return nil, errors.New("stsd box too short")
   }
   copy(b.HeaderFields[:], data[8:16])

   payload := data[16:b.Header.Size]
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
      case "encv", "enca":
         enc, err := DecodeEncBox(content)
         if err != nil {
            return nil, err
         }
         b.EncChildren = append(b.EncChildren, enc)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

func (b *StsdBox) Encode() []byte {
   buffer := make([]byte, 16)
   copy(buffer[8:16], b.HeaderFields[:])
   for _, child := range b.EncChildren {
      buffer = append(buffer, child.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *StsdBox) RemoveSinf() error {
   for _, child := range b.EncChildren {
      if child.Sinf == nil {
         continue
      }
      frma := child.Sinf.Frma
      if frma == nil {
         continue
      }
      child.Header.Type = frma.DataFormat
      child.Sinf = nil // Remove the sinf box
   }
   return nil
}

// --- TRAK ---
type TrakBox struct {
   Header      *BoxHeader
   Mdia        *MdiaBox
   RawChildren [][]byte
}

func DecodeTrakBox(data []byte) (*TrakBox, error) {
   b := &TrakBox{}
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
      case "mdia":
         mdia, err := DecodeMdiaBox(content)
         if err != nil {
            return nil, err
         }
         b.Mdia = mdia
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
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

// track.go
