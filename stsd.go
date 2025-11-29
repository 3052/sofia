package sofia

import (
   "encoding/binary"
   "errors"
)

type StsdChild struct {
   Enc *EncBox
   Raw []byte
}

type StsdBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StsdChild
}

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

      content := boxData[offset : offset+boxSize]
      var child StsdChild
      switch string(h.Type[:]) {
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
      offset += boxSize
   }
   return nil
}

func (b *StsdBox) Encode() []byte {
   buf := make([]byte, 8)

   // Copy raw fields (Version/Flags/Count) from original
   // 8 bytes: Header(8) -> [8:16] are Ver/Flags/EntryCount
   buf = append(buf, b.RawData[8:16]...)

   for _, child := range b.Children {
      if child.Enc != nil {
         buf = append(buf, child.Enc.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }

   b.Header.Size = uint32(len(buf))
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])
   return buf
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
