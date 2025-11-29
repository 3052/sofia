package sofia

type StblChild struct {
   Stsd *StsdBox
   Raw  []byte
}

type StblBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StblChild
}

func (b *StblBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child StblChild
      switch string(h.Type[:]) {
      case "stsd":
         var stsd StsdBox
         if err := stsd.Parse(content); err != nil {
            return err
         }
         child.Stsd = &stsd
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *StblBox) Encode() []byte {
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Stsd != nil {
         buf = append(buf, child.Stsd.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
   return buf
}

func (b *StblBox) Stsd() (*StsdBox, bool) {
   for _, child := range b.Children {
      if child.Stsd != nil {
         return child.Stsd, true
      }
   }
   return nil, false
}
