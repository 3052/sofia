package mp4

import "errors"

// SinfChild holds either a parsed box or raw data for a child of a 'sinf' box.
type SinfChild struct {
   Frma *FrmaBox
   Schi *SchiBox
   Raw  []byte
}

// SinfBox represents the 'sinf' box (Protection Scheme Information Box).
type SinfBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []SinfChild
}

// ParseSinf parses the 'sinf' box from a byte slice.
func ParseSinf(data []byte) (SinfBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SinfBox{}, err
   }
   var sinf SinfBox
   sinf.Header = header
   sinf.RawData = data[:header.Size]
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
         return SinfBox{}, errors.New("invalid child box size in sinf")
      }
      childData := boxData[offset : offset+boxSize]
      var child SinfChild
      switch string(h.Type[:]) {
      case "frma":
         frma, err := ParseFrma(childData)
         if err != nil {
            return SinfBox{}, err
         }
         child.Frma = &frma
      case "schi":
         schi, err := ParseSchi(childData)
         if err != nil {
            return SinfBox{}, err
         }
         child.Schi = &schi
      default:
         child.Raw = childData
      }
      sinf.Children = append(sinf.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return sinf, nil
}

// Encode re-serializes the box from its children.
func (b *SinfBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Frma != nil {
         content = append(content, child.Frma.Encode()...)
      } else if child.Schi != nil {
         content = append(content, child.Schi.Encode()...)
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
