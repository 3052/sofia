// config.go
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
   if len(data) < payloadOffset+entrySize {
      b.EntryHeader = data[payloadOffset:b.Header.Size]
      return b, nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+entrySize]

   payload := data[payloadOffset+entrySize : b.Header.Size]
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
      case "sinf":
         sinf, err := DecodeSinfBox(content)
         if err != nil {
            return nil, err
         }
         b.Sinf = sinf
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

func (b *EncBox) Encode() []byte {
   buffer := make([]byte, 8)
   buffer = append(buffer, b.EntryHeader...)
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

// --- STSD ---
type StsdBox struct {
   Header       *BoxHeader
   HeaderFields [8]byte // Ver(1)+Flags(3)+EntryCount(4)
   EncChildren  []*EncBox
   RawChildren  [][]byte
}

func (b *StsdBox) Sinf() (*SinfBox, *BoxHeader, bool) {
   for _, enc := range b.EncChildren {
      if enc.Sinf != nil {
         return enc.Sinf, enc.Header, true
      }
   }
   return nil, nil, false
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
