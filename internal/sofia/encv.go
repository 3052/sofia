package mp4

import "errors"

type EncvChild struct {
   Sinf *SinfBox
   Raw  []byte
}

type EncvBox struct {
   Header      BoxHeader
   RawData     []byte
   EntryHeader []byte
   Children    []EncvChild
}

func ParseEncv(data []byte) (EncvBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncvBox{}, err
   }
   var encv EncvBox
   encv.Header = header
   encv.RawData = data[:header.Size]
   const visualSampleEntrySize = 78
   payloadOffset := 8
   if len(data) < payloadOffset+visualSampleEntrySize {
      encv.EntryHeader = data[payloadOffset:header.Size]
      return encv, nil
   }
   encv.EntryHeader = data[payloadOffset : payloadOffset+visualSampleEntrySize]
   boxData := data[payloadOffset+visualSampleEntrySize : header.Size]
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
         return EncvBox{}, errors.New("invalid child box size in encv")
      }
      childData := boxData[offset : offset+boxSize]
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
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return encv, nil
}

func (b *EncvBox) Encode() []byte {
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
