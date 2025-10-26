package mp4

// MinfChild holds either a parsed box or raw data for a child of a 'minf' box.
type MinfChild struct {
   Stbl *StblBox
   Raw  []byte
}

// MinfBox represents the 'minf' box (Media Information Box).
type MinfBox struct {
   Header   BoxHeader
   Children []MinfChild
}

// ParseMinf parses the 'minf' box from a byte slice.
func ParseMinf(data []byte) (MinfBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MinfBox{}, err
   }
   var minf MinfBox
   minf.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MinfBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child MinfChild

      switch string(h.Type[:]) {
      case "stbl":
         stbl, err := ParseStbl(childData)
         if err != nil {
            return MinfBox{}, err
         }
         child.Stbl = &stbl
      default:
         child.Raw = childData
      }
      minf.Children = append(minf.Children, child)
      offset += int(h.Size)
   }
   return minf, nil
}

// Encode encodes the 'minf' box to a byte slice.
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
