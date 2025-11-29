package sofia

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
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child SinfChild
      switch string(h.Type[:]) {
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
