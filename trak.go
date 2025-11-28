package sofia

import "errors"

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
      case "mdia":
         var mdia MdiaBox
         if err := mdia.Parse(childData); err != nil {
            return err
         }
         child.Mdia = &mdia
      default:
         // 'edts' and others fall here
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
      if child.Mdia != nil {
         content = append(content, child.Mdia.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// Remove deletes all child boxes matching the given type (e.g., "edts").
func (b *TrakBox) Remove(boxType string) {
   var kept []TrakChild
   for _, child := range b.Children {
      // Check typed fields
      if boxType == "mdia" && child.Mdia != nil {
         continue // Skip to remove
      }
      // Check raw fields
      // Fix: Removed 'child.Raw != nil' check (S1009)
      if len(child.Raw) >= 8 {
         if string(child.Raw[4:8]) == boxType {
            continue // Skip to remove
         }
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
