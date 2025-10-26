package mp4

import "fmt"

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

func ParseTrak(data []byte) (TrakBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrakBox{}, err
   }
   var trak TrakBox
   trak.Header = header
   trak.RawData = data[:header.Size]
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return TrakBox{}, fmt.Errorf("invalid child box size in trak")
      }
      childData := boxData[offset : offset+boxSize]
      var child TrakChild
      switch string(h.Type[:]) {
      case "edts":
         edts, err := ParseEdts(childData)
         if err != nil {
            return TrakBox{}, err
         }
         child.Edts = &edts
      case "mdia":
         mdia, err := ParseMdia(childData)
         if err != nil {
            return TrakBox{}, err
         }
         child.Mdia = &mdia
      default:
         child.Raw = childData
      }
      trak.Children = append(trak.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return trak, nil
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
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
      var sinf *SinfBox
      if stsdChild.Encv != nil {
         for _, encvChild := range stsdChild.Encv.Children {
            if encvChild.Sinf != nil {
               sinf = encvChild.Sinf
               break
            }
         }
      }
      if sinf == nil && stsdChild.Enca != nil {
         for _, encaChild := range stsdChild.Enca.Children {
            if encaChild.Sinf != nil {
               sinf = encaChild.Sinf
               break
            }
         }
      }
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
