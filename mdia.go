package sofia

import "errors"

type MdiaChild struct {
   Mdhd *MdhdBox
   Minf *MinfBox
   Raw  []byte
}
type MdiaBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MdiaChild
}

func (b *MdiaBox) Parse(data []byte) error {
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
         return errors.New("invalid child box size in mdia")
      }
      childData := boxData[offset : offset+boxSize]
      var child MdiaChild
      switch string(h.Type[:]) {
      case "mdhd":
         var mdhd MdhdBox
         if err := mdhd.Parse(childData); err != nil {
            return err
         }
         child.Mdhd = &mdhd
      case "minf":
         var minf MinfBox
         if err := minf.Parse(childData); err != nil {
            return err
         }
         child.Minf = &minf
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
func (b *MdiaBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Mdhd != nil {
         content = append(content, child.Mdhd.Encode()...)
      } else if child.Minf != nil {
         content = append(content, child.Minf.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}
