package mp4

import "fmt"

// TrakChild holds either a parsed box or raw data for a child of a 'trak' box.
type TrakChild struct {
   Mdia *MdiaBox
   Raw  []byte
}

// TrakBox represents the 'trak' box (Track Box).
type TrakBox struct {
   Header   BoxHeader
   Children []TrakChild
}

// ParseTrak parses the 'trak' box from a byte slice.
func ParseTrak(data []byte) (TrakBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrakBox{}, err
   }
   var trak TrakBox
   trak.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return TrakBox{}, err
      }

      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 {
         return TrakBox{}, fmt.Errorf("invalid box size %d in trak", boxSize)
      }
      if offset+boxSize > len(boxData) {
         return TrakBox{}, fmt.Errorf("box size %d exceeds parent trak bounds", boxSize)
      }

      childData := boxData[offset : offset+boxSize]
      var child TrakChild

      switch string(h.Type[:]) {
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

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TrakBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Mdia != nil {
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

// --- Refactored Helper Functions ---

// GetStbl finds and returns the stbl box from within a trak box.
// This is a helper to reduce code duplication in other getters.
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

// GetStsd finds and returns the stsd box from within a trak box.
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

// GetTenc finds the tenc box by traversing the sample description.
// It reuses GetStsd to avoid duplicating traversal logic.
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
