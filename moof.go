package sofia

type MoofChild struct {
   Traf *TrafBox
   Pssh *PsshBox
   Raw  []byte
}

type MoofBox struct {
   Header   BoxHeader
   Children []MoofChild
}

func (b *MoofBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MoofChild
      switch string(h.Type[:]) {
      case "traf":
         var traf TrafBox
         if err := traf.Parse(content); err != nil {
            return err
         }
         child.Traf = &traf
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MoofBox) Traf() (*TrafBox, bool) {
   for _, child := range b.Children {
      if child.Traf != nil {
         return child.Traf, true
      }
   }
   return nil, false
}
