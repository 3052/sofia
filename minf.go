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

func (b *MinfBox) Parse(data []byte) error {
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
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}
