package sofia

type MinfChild struct {
   Stbl *StblBox
   Raw  []byte
}

type MinfBox struct {
   Header   BoxHeader
   Children []MinfChild
}

func (b *MinfBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MinfChild
      switch string(h.Type[:]) {
      case "stbl":
         var stbl StblBox
         if err := stbl.Parse(content); err != nil {
            return err
         }
         child.Stbl = &stbl
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MinfBox) Encode() []byte {
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Stbl != nil {
         buf = append(buf, child.Stbl.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
   return buf
}

func (b *MinfBox) Stbl() (*StblBox, bool) {
   for _, child := range b.Children {
      if child.Stbl != nil {
         return child.Stbl, true
      }
   }
   return nil, false
}
