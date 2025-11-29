package sofia

type TrakChild struct {
   Mdia *MdiaBox
   Raw  []byte
}

type TrakBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []TrakChild
}

func (b *TrakBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child TrakChild
      switch string(h.Type[:]) {
      case "mdia":
         var mdia MdiaBox
         if err := mdia.Parse(content); err != nil {
            return err
         }
         child.Mdia = &mdia
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *TrakBox) Encode() []byte {
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mdia != nil {
         buf = append(buf, child.Mdia.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
   return buf
}

func (b *TrakBox) RemoveEdts() {
   var kept []TrakChild
   for _, child := range b.Children {
      if len(child.Raw) >= 8 && string(child.Raw[4:8]) == "edts" {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *TrakBox) Mdia() (*MdiaBox, bool) {
   for _, child := range b.Children {
      if child.Mdia != nil {
         return child.Mdia, true
      }
   }
   return nil, false
}
