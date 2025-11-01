package sofia

import "errors"

type SchiChild struct {
   Tenc *TencBox
   Raw  []byte
}

type SchiBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []SchiChild
}

func (b *SchiBox) Parse(data []byte) error {
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
         return errors.New("invalid child box size in schi")
      }
      childData := boxData[offset : offset+boxSize]
      var child SchiChild
      switch string(h.Type[:]) {
      case "tenc":
         var tenc TencBox
         if err := tenc.Parse(childData); err != nil {
            return err
         }
         child.Tenc = &tenc
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

func (b *SchiBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Tenc != nil {
         content = append(content, child.Tenc.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// Tenc finds the TencBox child and returns it, along with a boolean indicating if it was found.
func (b *SchiBox) Tenc() (*TencBox, bool) {
   for _, child := range b.Children {
      if child.Tenc != nil {
         return child.Tenc, true
      }
   }
   return nil, false
}
