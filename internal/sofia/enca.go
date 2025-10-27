package mp4

import "errors"

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

// Parse parses the 'enca' box from a byte slice.
func (b *EncaBox) Parse(data []byte) error {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return err
   }
   b.Header = header
   b.RawData = data[:header.Size]
   const audioSampleEntrySize = 28
   payloadOffset := 8
   if len(data) < payloadOffset+audioSampleEntrySize {
      b.EntryHeader = data[payloadOffset:header.Size]
      return nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+audioSampleEntrySize]
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
         return errors.New("invalid child box size in enca")
      }
      childData := boxData[offset : offset+boxSize]
      var child EncaChild
      switch string(h.Type[:]) {
      case "sinf":
         var sinf SinfBox
         if err := sinf.Parse(childData); err != nil {
            return err
         }
         child.Sinf = &sinf
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
