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

func ParseMdia(data []byte) (MdiaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdiaBox{}, err
   }
   var mdia MdiaBox
   mdia.Header = header
   mdia.RawData = data[:header.Size]
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
         return MdiaBox{}, errors.New("invalid child box size in mdia")
      }
      childData := boxData[offset : offset+boxSize]
      var child MdiaChild
      switch string(h.Type[:]) {
      case "mdhd":
         mdhd, err := ParseMdhd(childData)
         if err != nil {
            return MdiaBox{}, err
         }
         child.Mdhd = &mdhd
      case "minf":
         minf, err := ParseMinf(childData)
         if err != nil {
            return MdiaBox{}, err
         }
         child.Minf = &minf
      default:
         child.Raw = childData
      }
      mdia.Children = append(mdia.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return mdia, nil
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
