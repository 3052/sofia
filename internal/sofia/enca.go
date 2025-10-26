package mp4

import "fmt"

type EncaChild struct {
   Sinf *SinfBox
   Raw  []byte
}
type EncaBox struct {
   Header      BoxHeader
   RawData     []byte
   EntryHeader []byte
   Children    []EncaChild
}

func ParseEnca(data []byte) (EncaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncaBox{}, err
   }
   var enca EncaBox
   enca.Header = header
   enca.RawData = data[:header.Size]
   const audioSampleEntrySize = 28
   payloadOffset := 8
   if len(data) < payloadOffset+audioSampleEntrySize {
      enca.EntryHeader = data[payloadOffset:header.Size]
      return enca, nil
   }
   enca.EntryHeader = data[payloadOffset : payloadOffset+audioSampleEntrySize]
   boxData := data[payloadOffset+audioSampleEntrySize : header.Size]
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
         return EncaBox{}, fmt.Errorf("invalid child box size in enca")
      }
      childData := boxData[offset : offset+boxSize]
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
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return enca, nil
}
func (b *EncaBox) Encode() []byte {
   var childrenContent []byte
   for _, child := range b.Children {
      if child.Sinf != nil {
         childrenContent = append(childrenContent, child.Sinf.Encode()...)
      } else if child.Raw != nil {
         childrenContent = append(childrenContent, child.Raw...)
      }
   }
   content := append(b.EntryHeader, childrenContent...)
   b.Header.Size = uint32(8 + len(content))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
