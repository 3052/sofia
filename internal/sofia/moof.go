package mp4

import (
   "errors"
)

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

// Parse parses the 'moof' box from a byte slice.
func (b *MoofBox) Parse(data []byte) error {
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
func (b *MoofBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Traf != nil {
         content = append(content, child.Traf.Encode()...)
      } else if child.Pssh != nil {
         content = append(content, child.Pssh.Encode()...)
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

// Sanitize finds and renames all pssh boxes within this moof box to 'free'.
func (b *MoofBox) Sanitize() {
   for i := range b.Children {
      child := &b.Children[i]
      if child.Pssh != nil {
         child.Pssh.Header.Type = [4]byte{'f', 'r', 'e', 'e'}
      }
   }
}
