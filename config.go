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

func (b *StsdBox) Sinf() (*SinfBox, *BoxHeader, bool) {
   for _, enc := range b.EncChildren {
      if enc.Sinf != nil {
         return enc.Sinf, enc.Header, true
      }
   }
   return nil, nil, false
}

// config.go
