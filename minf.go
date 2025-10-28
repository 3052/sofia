package sofia

import "errors"

type MinfChild struct {
   Stbl *StblBox
   Raw  []byte
}

type MinfBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MinfChild
}

// Parse parses the 'minf' box from a byte slice.
func (b *MinfBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if _, err := h.Read(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in minf")
      }
      childData := boxData[offset : offset+boxSize]
      var child MinfChild
      switch string(h.Type[:]) {
      case "stbl":
         var stbl StblBox
         if err := stbl.Parse(childData); err != nil {
            return err
         }
         child.Stbl = &stbl
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

func (b *MinfBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Stbl != nil {
         content = append(content, child.Stbl.Encode()...)
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
