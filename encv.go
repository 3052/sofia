package sofia

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

func (b *EncvBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   const visualSampleEntrySize = 78
   payloadOffset := 8
   if len(data) < payloadOffset+visualSampleEntrySize {
      b.EntryHeader = data[payloadOffset:b.Header.Size]
      return nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+visualSampleEntrySize]
   boxData := data[payloadOffset+visualSampleEntrySize : b.Header.Size]
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
         return errors.New("invalid child box size in encv")
      }
      childData := boxData[offset : offset+boxSize]
      var child EncvChild
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
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}
