package mp4

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

func ParseStbl(data []byte) (StblBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StblBox{}, err
   }
   var stbl StblBox
   stbl.Header = header
   stbl.RawData = data[:header.Size]
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
         return StblBox{}, errors.New("invalid child box size in stbl")
      }
      childData := boxData[offset : offset+boxSize]
      var child StblChild
      switch string(h.Type[:]) {
      case "stsd":
         stsd, err := ParseStsd(childData)
         if err != nil {
            return StblBox{}, err
         }
         child.Stsd = &stsd
      default:
         child.Raw = childData
      }
      stbl.Children = append(stbl.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return stbl, nil
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
