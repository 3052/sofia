package mp4

// StblChild holds either a parsed box or raw data for a child of an 'stbl' box.
type StblChild struct {
   Stsd *StsdBox
   Raw  []byte
}

// StblBox represents the 'stbl' box (Sample Table Box).
type StblBox struct {
   Header   BoxHeader
   Children []StblChild
}

// ParseStbl parses the 'stbl' box from a byte slice.
func ParseStbl(data []byte) (StblBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StblBox{}, err
   }
   var stbl StblBox
   stbl.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return StblBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
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
      offset += int(h.Size)
   }
   return stbl, nil
}

// Encode encodes the 'stbl' box to a byte slice.
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
