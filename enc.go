package sofia

import "fmt"

type EncChild struct {
   Sinf *SinfBox
   Raw  []byte
}

type EncBox struct {
   Header      BoxHeader
   EntryHeader []byte
   Children    []EncChild
}

func (b *EncBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

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

   return parseContainer(data[payloadOffset+entrySize:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child EncChild
      switch string(h.Type[:]) {
      case "sinf":
         var sinf SinfBox
         if err := sinf.Parse(content); err != nil {
            return err
         }
         child.Sinf = &sinf
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *EncBox) Encode() []byte {
   buf := make([]byte, 8)
   buf = append(buf, b.EntryHeader...)

   for _, child := range b.Children {
      // skip sinf
      if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
   return buf
}

func (b *EncBox) Unprotect() error {
   var sinf *SinfBox
   kept := make([]EncChild, 0, len(b.Children))

   for _, child := range b.Children {
      if child.Sinf != nil {
         if sinf == nil {
            sinf = child.Sinf
         }
         continue
      }
      kept = append(kept, child)
   }

   if sinf == nil {
      return nil
   }

   frma := sinf.Frma()
   if frma == nil {
      // handle edge case
      return nil
   }

   b.Header.Type = frma.DataFormat
   b.Children = kept

   return nil
}
