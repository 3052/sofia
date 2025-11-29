package sofia

import (
   "encoding/binary"
   "errors"
)

type StsdChild struct {
   Enc *EncBox // Handles both enca and encv
   Raw []byte
}

type StsdBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StsdChild
}

// Sinf finds the first protected sample entry and returns its SinfBox.
func (b *StsdBox) Sinf() (*SinfBox, *BoxHeader, bool) {
   for i := range b.Children {
      child := &b.Children[i]
      if child.Enc != nil {
         for _, c := range child.Enc.Children {
            if c.Sinf != nil {
               return c.Sinf, &child.Enc.Header, true
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
   b.RawData = data[:b.Header.Size]
   entryCount := binary.BigEndian.Uint32(data[12:16])
   boxData := data[16:b.Header.Size]
   offset := 0
   for i := uint32(0); i < entryCount && offset < len(boxData); i++ {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in stsd")
      }
      childData := boxData[offset : offset+boxSize]
      var child StsdChild
      switch string(h.Type[:]) {
      case "encv", "enca":
         var enc EncBox
         if err := enc.Parse(childData); err != nil {
            return err
         }
         child.Enc = &enc
      default:
         child.Raw = childData
      }
      b.Children = append(b.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return nil
}

func (b *StsdBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Enc != nil {
         content = append(content, child.Enc.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   headerData := b.RawData[8:16]
   fullContent := append(headerData, content...)
   b.Header.Size = uint32(8 + len(fullContent))
   headerBytes := b.Header.Encode()
   return append(headerBytes, fullContent...)
}

// UnprotectAll iterates over all sample entries and unprotects them
// if they are encrypted (enca/encv).
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
