package sofia

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

func (b *EncaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   const audioSampleEntrySize = 28
   payloadOffset := 8
   if len(data) < payloadOffset+audioSampleEntrySize {
      b.EntryHeader = data[payloadOffset:b.Header.Size]
      return nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+audioSampleEntrySize]
   boxData := data[payloadOffset+audioSampleEntrySize : b.Header.Size]
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
      // Logic update: Skip 'sinf' entirely to delete it.
      // We only write back 'Raw' children.
      if child.Raw != nil {
         childrenContent = append(childrenContent, child.Raw...)
      }
   }
   content := append(b.EntryHeader, childrenContent...)
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}
