package mp4

import (
   "encoding/binary"
   "errors"
)

type StsdChild struct {
   Encv *EncvBox
   Enca *EncaBox
   Raw  []byte
}
type StsdBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []StsdChild
}

// Parse parses the 'stsd' box from a byte slice.
func (b *StsdBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   entryCount := binary.BigEndian.Uint32(data[12:16])
   boxData := data[16:b.Header.Size]
   offset := 0
   for i := uint32(0); i < entryCount && offset < len(boxData); i++ {
      var h BoxHeader
      if _, err := h.Read(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in stsd")
      }
      childData := boxData[offset : offset+boxSize]
      var child StsdChild
      switch string(h.Type[:]) {
      case "encv":
         var encv EncvBox
         if err := encv.Parse(childData); err != nil {
            return err
         }
         child.Encv = &encv
      case "enca":
         var enca EncaBox
         if err := enca.Parse(childData); err != nil {
            return err
         }
         child.Enca = &enca
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
func (b *StsdBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Encv != nil {
         content = append(content, child.Encv.Encode()...)
      } else if child.Enca != nil {
         content = append(content, child.Enca.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   headerData := b.RawData[8:16] // Get original version, flags, and entry count
   fullContent := append(headerData, content...)
   b.Header.Size = uint32(8 + len(fullContent))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], fullContent)
   return encoded
}
