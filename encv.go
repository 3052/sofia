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
      // Do NOT encode child.Sinf. It is read-only/deleted.
      if child.Raw != nil {
         childrenContent = append(childrenContent, child.Raw...)
      }
   }
   content := append(b.EntryHeader, childrenContent...)
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

// Remove deletes children matching the box type (e.g. "sinf").
func (b *EncvBox) Remove(boxType string) {
   var kept []EncvChild
   for _, child := range b.Children {
      if boxType == "sinf" && child.Sinf != nil {
         continue
      }
      // Fix: Removed 'child.Raw != nil' check (S1009)
      if len(child.Raw) >= 8 {
         if string(child.Raw[4:8]) == boxType {
            continue
         }
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

// Unprotect converts this 'encv' box into a cleartext visual sample entry (e.g. 'avc1').
func (b *EncvBox) Unprotect() error {
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

   // 1. Change Header Type to original format (e.g. "avc1")
   b.Header.Type = frma.DataFormat

   // 2. Remove 'sinf'
   b.Remove("sinf")

   return nil
}
