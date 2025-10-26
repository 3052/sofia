package mp4

import "fmt"

type SchiChild struct {
   Tenc *TencBox
   Raw  []byte
}
type SchiBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []SchiChild
}

func ParseSchi(data []byte) (SchiBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SchiBox{}, err
   }
   var schi SchiBox
   schi.Header = header
   schi.RawData = data[:header.Size]
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
         return SchiBox{}, fmt.Errorf("invalid child box size in schi")
      }
      childData := boxData[offset : offset+boxSize]
      var child SchiChild
      switch string(h.Type[:]) {
      case "tenc":
         tenc, err := ParseTenc(childData)
         if err != nil {
            return SchiBox{}, err
         }
         child.Tenc = &tenc
      default:
         child.Raw = childData
      }
      schi.Children = append(schi.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return schi, nil
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
