package mp4

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

// Parse parses the 'mdia' box from a byte slice.
func (b *MdiaBox) Parse(data []byte) error {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return err
   }
   b.Header = header
   b.RawData = data[:header.Size]
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
