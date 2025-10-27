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

// Parse parses the 'sinf' box from a byte slice.
func (b *SinfBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if _, err := h.Read(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in sinf")
      }
      childData := boxData[offset : offset+boxSize]
      var child SinfChild
      switch string(h.Type[:]) {
      case "frma":
         var frma FrmaBox
         if err := frma.Parse(childData); err != nil {
            return err
         }
         child.Frma = &frma
      case "schi":
         var schi SchiBox
         if err := schi.Parse(childData); err != nil {
            return err
         }
         child.Schi = &schi
      default:
         child.Raw = childData
      }
      b.Children = append(b.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return nil
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
