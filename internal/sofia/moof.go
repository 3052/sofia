package mp4

// MoofChild holds either a parsed box or raw data for a child of a 'moof' box.
type MoofChild struct {
   Traf *TrafBox
   Raw  []byte
}

// MoofBox represents the 'moof' box (Movie Fragment Box).
type MoofBox struct {
   Header   BoxHeader
   Children []MoofChild
}

// ParseMoof parses the 'moof' box from a byte slice.
func ParseMoof(data []byte) (MoofBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoofBox{}, err
   }
   var moof MoofBox
   moof.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MoofBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child MoofChild

      switch string(h.Type[:]) {
      case "traf":
         traf, err := ParseTraf(childData)
         if err != nil {
            return MoofBox{}, err
         }
         child.Traf = &traf
      default:
         child.Raw = childData
      }
      moof.Children = append(moof.Children, child)
      offset += int(h.Size)
   }
   return moof, nil
}

// Encode encodes the 'moof' box to a byte slice.
func (b *MoofBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Traf != nil {
         content = append(content, child.Traf.Encode()...)
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
