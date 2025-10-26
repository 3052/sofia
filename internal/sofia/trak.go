package mp4

// TrakChild holds either a parsed box or raw data for a child of a 'trak' box.
type TrakChild struct {
   Mdia *MdiaBox
   Raw  []byte
}

// TrakBox represents the 'trak' box (Track Box).
type TrakBox struct {
   Header   BoxHeader
   Children []TrakChild
}

// ParseTrak parses the 'trak' box from a byte slice.
func ParseTrak(data []byte) (TrakBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrakBox{}, err
   }
   var trak TrakBox
   trak.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return TrakBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child TrakChild

      switch string(h.Type[:]) {
      case "mdia":
         mdia, err := ParseMdia(childData)
         if err != nil {
            return TrakBox{}, err
         }
         child.Mdia = &mdia
      default:
         child.Raw = childData
      }
      trak.Children = append(trak.Children, child)
      offset += int(h.Size)
   }
   return trak, nil
}

// Encode encodes the 'trak' box to a byte slice.
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
