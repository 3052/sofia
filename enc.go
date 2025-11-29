package sofia

import (
   "encoding/binary"
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
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])
   return buf
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
      return errors.New("cannot unprotect: sinf missing frma")
   }

   b.Header.Type = frma.DataFormat

   var kept []EncChild
   for _, child := range b.Children {
      if child.Sinf == nil {
         kept = append(kept, child)
      }
   }
   b.Children = kept

   return nil
}
