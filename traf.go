package sofia

type TrafChild struct {
   Tfhd *TfhdBox
   Trun *TrunBox
   Senc *SencBox
   Raw  []byte
}

type TrafBox struct {
   Header   BoxHeader
   Children []TrafChild
}

func (b *TrafBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child TrafChild
      switch string(h.Type[:]) {
      case "tfhd":
         var tfhd TfhdBox
         if err := tfhd.Parse(content); err != nil {
            return err
         }
         child.Tfhd = &tfhd
      case "trun":
         var trun TrunBox
         if err := trun.Parse(content); err != nil {
            return err
         }
         child.Trun = &trun
      case "senc":
         var senc SencBox
         if err := senc.Parse(content); err != nil {
            return err
         }
         child.Senc = &senc
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *TrafBox) Tfhd() *TfhdBox {
   for _, child := range b.Children {
      if child.Tfhd != nil {
         return child.Tfhd
      }
   }
   return nil
}

func (b *TrafBox) Trun() *TrunBox {
   for _, child := range b.Children {
      if child.Trun != nil {
         return child.Trun
      }
   }
   return nil
}

func (b *TrafBox) Senc() (*SencBox, bool) {
   for _, child := range b.Children {
      if child.Senc != nil {
         return child.Senc, true
      }
   }
   return nil, false
}
