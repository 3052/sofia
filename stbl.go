package sofia

import "errors"

type StblChild struct {
   Stsd *StsdBox
   Raw  []byte
}

type StblBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StblChild
}

func (b *StblBox) Parse(data []byte) error {
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
         return errors.New("invalid child box size in stbl")
      }
      childData := boxData[offset : offset+boxSize]
      var child StblChild
      switch string(h.Type[:]) {
      case "stsd":
         var stsd StsdBox
         if err := stsd.Parse(childData); err != nil {
            return err
         }
         child.Stsd = &stsd
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

func (b *StblBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Stsd != nil {
         content = append(content, child.Stsd.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// Stsd finds the StsdBox child and returns it, along with a boolean indicating if it was found.
func (b *StblBox) Stsd() (*StsdBox, bool) {
   for _, child := range b.Children {
      if child.Stsd != nil {
         return child.Stsd, true
      }
   }
   return nil, false
}
