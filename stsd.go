package sofia

import "errors"

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
   if len(data) < 16 {
      return errors.New("stsd box too short")
   }
   // Copy Version(1) + Flags(3) + EntryCount(4)
   copy(b.HeaderFields[:], data[8:16])

   // Parse children starting at offset 16
   return parseContainer(data[16:b.Header.Size], func(h BoxHeader, content []byte) error {
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
      return nil
   })
}

func (b *StsdBox) Encode() []byte {
   // Header(8) + HeaderFields(8)
   buf := make([]byte, 16)
   copy(buf[8:16], b.HeaderFields[:])

   for _, child := range b.Children {
      if child.Enc != nil {
         buf = append(buf, child.Enc.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
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
