package sofia

import "errors"

type TrakChild struct {
   Edts *EdtsBox
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
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in trak")
      }
      childData := boxData[offset : offset+boxSize]
      var child TrakChild
      switch string(h.Type[:]) {
      case "edts":
         var edts EdtsBox
         if err := edts.Parse(childData); err != nil {
            return err
         }
         child.Edts = &edts
      case "mdia":
         var mdia MdiaBox
         if err := mdia.Parse(childData); err != nil {
            return err
         }
         child.Mdia = &mdia
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
func (b *TrakBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Edts != nil {
         content = append(content, child.Edts.Encode()...)
      } else if child.Mdia != nil {
         content = append(content, child.Mdia.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

func (b *TrakBox) RemoveEdts() {
   for i := range b.Children {
      child := &b.Children[i]
      if child.Edts != nil {
         child.Edts.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
      }
   }
}

// GetMdhd finds the MdhdBox and returns it, along with a boolean indicating if it was found.
func (b *TrakBox) GetMdhd() (*MdhdBox, bool) {
   for _, child := range b.Children {
      if mdia := child.Mdia; mdia != nil {
         for _, mdiaChild := range mdia.Children {
            if mdhd := mdiaChild.Mdhd; mdhd != nil {
               return mdhd, true
            }
         }
      }
   }
   return nil, false
}
func (b *TrakBox) GetStbl() *StblBox {
   for _, child := range b.Children {
      if mdia := child.Mdia; mdia != nil {
         for _, mdiaChild := range mdia.Children {
            if minf := mdiaChild.Minf; minf != nil {
               for _, minfChild := range minf.Children {
                  if stbl := minfChild.Stbl; stbl != nil {
                     return stbl
                  }
               }
            }
         }
      }
   }
   return nil
}
func (b *TrakBox) GetStsd() *StsdBox {
   stbl := b.GetStbl()
   if stbl == nil {
      return nil
   }
   for _, stblChild := range stbl.Children {
      if stsd := stblChild.Stsd; stsd != nil {
         return stsd
      }
   }
   return nil
}
func (b *TrakBox) GetTenc() *TencBox {
   stsd := b.GetStsd()
   if stsd == nil {
      return nil
   }
   for _, stsdChild := range stsd.Children {
      sinf := stsdChild.GetSinf()
      if sinf != nil {
         for _, sinfChild := range sinf.Children {
            if schi := sinfChild.Schi; schi != nil {
               for _, schiChild := range schi.Children {
                  if schiChild.Tenc != nil {
                     return schiChild.Tenc
                  }
               }
            }
         }
      }
   }
   return nil
}
