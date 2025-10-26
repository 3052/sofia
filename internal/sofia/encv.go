package mp4

// EncvChild holds either a parsed box or raw data for a child of an 'encv' box.
type EncvChild struct {
   Sinf *SinfBox
   Raw  []byte
}

// EncvBox represents the 'encv' box (Encrypted Video).
type EncvBox struct {
   Header   BoxHeader
   Children []EncvChild
}

// ParseEncv parses the 'encv' box from a byte slice.
func ParseEncv(data []byte) (EncvBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncvBox{}, err
   }
   var encv EncvBox
   encv.Header = header
   boxData := data[8:header.Size]

   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return EncvBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child EncvChild

      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(childData)
         if err != nil {
            return EncvBox{}, err
         }
         child.Sinf = &sinf
      default:
         child.Raw = childData
      }
      encv.Children = append(encv.Children, child)
      offset += int(h.Size)
   }
   return encv, nil
}

// Encode encodes the 'encv' box to a byte slice.
func (b *EncvBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Sinf != nil {
         content = append(content, child.Sinf.Encode()...)
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
