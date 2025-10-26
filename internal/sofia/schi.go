package mp4

// SchiChild holds either a parsed box or raw data for a child of a 'schi' box.
type SchiChild struct {
   Tenc *TencBox
   Raw  []byte
}

// SchiBox represents the 'schi' box (Scheme Information Box).
type SchiBox struct {
   Header   BoxHeader
   Children []SchiChild
}

// ParseSchi parses the 'schi' box from a byte slice.
func ParseSchi(data []byte) (SchiBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SchiBox{}, err
   }
   var schi SchiBox
   schi.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return SchiBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
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
      offset += int(h.Size)
   }
   return schi, nil
}

// Encode encodes the 'schi' box to a byte slice.
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
