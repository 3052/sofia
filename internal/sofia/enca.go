package mp4

// EncaChild holds either a parsed box or raw data for a child of an 'enca' box.
type EncaChild struct {
   Sinf *SinfBox
   Raw  []byte
}

// EncaBox represents the 'enca' box (Encrypted Audio).
type EncaBox struct {
   Header   BoxHeader
   Children []EncaChild
}

// ParseEnca parses the 'enca' box from a byte slice.
func ParseEnca(data []byte) (EncaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncaBox{}, err
   }
   var enca EncaBox
   enca.Header = header
   boxData := data[8:header.Size]

   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return EncaBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child EncaChild

      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(childData)
         if err != nil {
            return EncaBox{}, err
         }
         child.Sinf = &sinf
      default:
         child.Raw = childData
      }
      enca.Children = append(enca.Children, child)
      offset += int(h.Size)
   }
   return enca, nil
}

// Encode encodes the 'enca' box to a byte slice.
func (b *EncaBox) Encode() []byte {
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
