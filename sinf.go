package sofia

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
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
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
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// GetFrma finds and returns the first FrmaBox child.
func (b *SinfBox) GetFrma() *FrmaBox {
   for _, child := range b.Children {
      if child.Frma != nil {
         return child.Frma
      }
   }
   return nil
}
