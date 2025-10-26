package mp4

import (
   "encoding/binary"
   "fmt"
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

func ParseStsd(data []byte) (StsdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StsdBox{}, err
   }
   var stsd StsdBox
   stsd.Header = header
   stsd.RawData = data[:header.Size]
   entryCount := binary.BigEndian.Uint32(data[12:16])
   boxData := data[16:header.Size]
   offset := 0
   for i := uint32(0); i < entryCount && offset < len(boxData); i++ {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return StsdBox{}, fmt.Errorf("invalid child box size in stsd")
      }
      childData := boxData[offset : offset+boxSize]
      var child StsdChild
      switch string(h.Type[:]) {
      case "encv":
         encv, err := ParseEncv(childData)
         if err != nil {
            return StsdBox{}, err
         }
         child.Encv = &encv
      case "enca":
         enca, err := ParseEnca(childData)
         if err != nil {
            return StsdBox{}, err
         }
         child.Enca = &enca
      default:
         child.Raw = childData
      }
      stsd.Children = append(stsd.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return stsd, nil
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
