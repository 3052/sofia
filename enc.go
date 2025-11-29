package sofia

import (
   "errors"
   "fmt"
)

type EncChild struct {
   Sinf *SinfBox
   Raw  []byte
}

type EncBox struct {
   Header      BoxHeader
   RawData     []byte
   EntryHeader []byte
   Children    []EncChild
}

func (b *EncBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]

   // Determine entry header size based on box type
   var entrySize int
   switch string(b.Header.Type[:]) {
   case "enca":
      entrySize = 28
   case "encv":
      entrySize = 78
   default:
      return fmt.Errorf("unknown encryption box type: %s", string(b.Header.Type[:]))
   }

   payloadOffset := 8
   if len(data) < payloadOffset+entrySize {
      b.EntryHeader = data[payloadOffset:b.Header.Size]
      return nil
   }
   b.EntryHeader = data[payloadOffset : payloadOffset+entrySize]

   boxData := data[payloadOffset+entrySize : b.Header.Size]
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
         return errors.New("invalid child box size in encrypted entry")
      }
      childData := boxData[offset : offset+boxSize]
      var child EncChild
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

func (b *EncBox) Encode() []byte {
   var childrenContent []byte
   for _, child := range b.Children {
      // sinf is skipped (read-only/deleted)
      if child.Raw != nil {
         childrenContent = append(childrenContent, child.Raw...)
      }
   }
   content := append(b.EntryHeader, childrenContent...)
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

func (b *EncBox) Unprotect() error {
   var sinf *SinfBox
   for _, child := range b.Children {
      if child.Sinf != nil {
         sinf = child.Sinf
         break
      }
   }
   if sinf == nil {
      return nil
   }

   frma := sinf.Frma()
   if frma == nil {
      return errors.New("cannot unprotect: sinf box missing frma")
   }

   // Change Type to original format (e.g. "avc1" or "mp4a")
   b.Header.Type = frma.DataFormat

   // Remove 'sinf' child
   var kept []EncChild
   for _, child := range b.Children {
      if child.Sinf != nil {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept

   return nil
}
