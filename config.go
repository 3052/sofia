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
type StsdChild struct {
   Enc *EncBox
   Raw []byte
}

type StsdBox struct {
   Header       BoxHeader
   HeaderFields [8]byte // Ver(1)+Flags(3)+EntryCount(4)
   Children     []StsdChild
}

func (b *StsdBox) Sinf() (*SinfBox, *BoxHeader, bool) {
   for i := range b.Children {
      child := &b.Children[i]
      if child.Enc != nil {
         for _, encChild := range child.Enc.Children {
            if encChild.Sinf != nil {
               return encChild.Sinf, &child.Enc.Header, true
            }
         }
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
   // Copy Version(1) + Flags(3) + EntryCount(4)
   copy(b.HeaderFields[:], data[8:16])
   // Parse children starting at offset 16
   return parseContainer(data[16:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child StsdChild
      switch string(header.Type[:]) {
      case "encv", "enca":
         var enc EncBox
         if err := enc.Parse(content); err != nil {
            return err
         }
         child.Enc = &enc
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *StsdBox) Encode() []byte {
   // Header(8) + HeaderFields(8)
   buffer := make([]byte, 16)
   copy(buffer[8:16], b.HeaderFields[:])
   for _, child := range b.Children {
      if child.Enc != nil {
         buffer = append(buffer, child.Enc.Encode()...)
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *StsdBox) UnprotectAll() error {
   for _, child := range b.Children {
      if child.Enc != nil {
         if err := child.Enc.Unprotect(); err != nil {
            return err
         }
      }
   }
   return nil
}

// --- ENC (Encrypted Sample Entry) ---
type EncChild struct {
   Sinf *SinfBox
   Raw  []byte
}

type EncBox struct {
   Header      BoxHeader
   EntryHeader []byte
   Children    []EncChild
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
   return parseContainer(data[payloadOffset+entrySize:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child EncChild
      switch string(header.Type[:]) {
      case "sinf":
         var sinf SinfBox
         if err := sinf.Parse(content); err != nil {
            return err
         }
         child.Sinf = &sinf
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *EncBox) Encode() []byte {
   buffer := make([]byte, 8)
   buffer = append(buffer, b.EntryHeader...)
   for _, child := range b.Children {
      // skip sinf
      if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *EncBox) Unprotect() error {
   var sinf *SinfBox
   kept := make([]EncChild, 0, len(b.Children))
   for _, child := range b.Children {
      if child.Sinf != nil {
         if sinf == nil {
            sinf = child.Sinf
         }
         continue
      }
      kept = append(kept, child)
   }
   if sinf == nil {
      return nil
   }
   frma := sinf.Frma()
   if frma == nil {
      // handle edge case
      return nil
   }
   b.Header.Type = frma.DataFormat
   b.Children = kept
   return nil
}

// --- SINF ---
type SinfChild struct {
   Frma *FrmaBox
   Raw  []byte
}

type SinfBox struct {
   Header   BoxHeader
   Children []SinfChild
}

func (b *SinfBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child SinfChild
      switch string(header.Type[:]) {
      case "frma":
         var frma FrmaBox
         if err := frma.Parse(content); err != nil {
            return err
         }
         child.Frma = &frma
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *SinfBox) Frma() *FrmaBox {
   for _, child := range b.Children {
      if child.Frma != nil {
         return child.Frma
      }
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
