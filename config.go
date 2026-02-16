package sofia

import (
   "errors"
   "strings"
)

func encryptionBoxError(boxType [4]byte) error {
   var sb strings.Builder
   sb.WriteString("unknown encryption box type ")
   sb.Write(boxType[:])
   return errors.New(sb.String())
}

// --- STSD ---
type StsdBox struct {
   Header       BoxHeader
   HeaderFields [8]byte // Ver(1)+Flags(3)+EntryCount(4)
   EncChildren  []*EncBox
   RawChildren  [][]byte
}

func (b *StsdBox) Sinf() (*SinfBox, *BoxHeader, bool) {
   for _, enc := range b.EncChildren {
      if enc.Sinf != nil {
         return enc.Sinf, &enc.Header, true
      }
   }
   return nil, nil, false
}

func (b *StsdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 16 {
      return errors.New("stsd box too short")
   }
   copy(b.HeaderFields[:], data[8:16])

   payload := data[16:b.Header.Size]
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
      case "encv", "enca":
         var enc EncBox
         if err := enc.Parse(content); err != nil {
            return err
         }
         b.EncChildren = append(b.EncChildren, &enc)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
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

func (b *StsdBox) UnprotectAll() error {
   for _, child := range b.EncChildren {
      if err := child.Unprotect(); err != nil {
         return err
      }
   }
   return nil
}

// --- ENC (Encrypted Sample Entry) ---
type EncBox struct {
   Header      BoxHeader
   EntryHeader []byte
   Sinf        *SinfBox
   RawChildren [][]byte
}

func (b *EncBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   var entrySize int
   switch string(b.Header.Type[:]) {
   case "enca":
      entrySize = 28
   case "encv":
      entrySize = 78
   default:
      return encryptionBoxError(b.Header.Type)
   }
   payloadOffset := 8
   if len(data) < payloadOffset+entrySize {
      b.EntryHeader = data[payloadOffset:b.Header.Size]
      return nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+entrySize]

   payload := data[payloadOffset+entrySize : b.Header.Size]
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
      case "sinf":
         var sinf SinfBox
         if err := sinf.Parse(content); err != nil {
            return err
         }
         b.Sinf = &sinf
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
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

func (b *EncBox) Unprotect() error {
   if b.Sinf == nil {
      return nil
   }
   frma := b.Sinf.Frma
   if frma == nil {
      return nil
   }
   b.Header.Type = frma.DataFormat
   b.Sinf = nil // Remove the sinf box
   return nil
}

// --- SINF ---
type SinfBox struct {
   Header      BoxHeader
   Frma        *FrmaBox
   Schi        *SchiBox
   RawChildren [][]byte
}

func (b *SinfBox) Parse(data []byte) error {
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
      case "frma":
         var frma FrmaBox
         if err := frma.Parse(content); err != nil {
            return err
         }
         b.Frma = &frma
      case "schi":
         var schi SchiBox
         if err := schi.Parse(content); err != nil {
            return err
         }
         b.Schi = &schi
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

// --- SCHI (Scheme Information) ---
type SchiBox struct {
   Header      BoxHeader
   Tenc        *TencBox
   RawChildren [][]byte
}

func (b *SchiBox) Parse(data []byte) error {
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
      case "tenc":
         var tenc TencBox
         if err := tenc.Parse(content); err != nil {
            return err
         }
         b.Tenc = &tenc
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

// --- FRMA ---
type FrmaBox struct {
   Header     BoxHeader
   DataFormat [4]byte
}

func (b *FrmaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("frma box is too small")
   }
   copy(b.DataFormat[:], data[8:12])
   return nil
}
