package sofia

import "errors"

type MoofChild struct {
   Traf *TrafBox
   Pssh *PsshBox
   Raw  []byte
}

type MoofBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MoofChild
}

func (b *MoofBox) Parse(data []byte) error {
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
         return errors.New("invalid child box size in moof")
      }
      childData := boxData[offset : offset+boxSize]
      var child MoofChild
      switch string(h.Type[:]) {
      case "traf":
         var traf TrafBox
         if err := traf.Parse(childData); err != nil {
            return err
         }
         child.Traf = &traf
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(childData); err != nil {
            return err
         }
         child.Pssh = &pssh
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

// Traf returns the first traf box found and a boolean indicating if it was found.
func (b *MoofBox) Traf() (*TrafBox, bool) {
   for _, child := range b.Children {
      if child.Traf != nil {
         return child.Traf, true
      }
   }
   return nil, false
}
