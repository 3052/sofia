package mp4

// MdiaChild holds either a parsed box or raw data for a child of an 'mdia' box.
type MdiaChild struct {
   Mdhd *MdhdBox
   Minf *MinfBox
   Raw  []byte
}

// MdiaBox represents the 'mdia' box (Media Box).
type MdiaBox struct {
   Header   BoxHeader
   Children []MdiaChild
}

// ParseMdia parses the 'mdia' box from a byte slice.
func ParseMdia(data []byte) (MdiaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdiaBox{}, err
   }
   var mdia MdiaBox
   mdia.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MdiaBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
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
      offset += int(h.Size)
   }
   return mdia, nil
}

// Encode encodes the 'mdia' box to a byte slice.
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
