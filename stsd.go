package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class SampleDescriptionBox() extends FullBox(
//    'stsd',
//    version,
//    0
// ) {
//    int i ;
//    unsigned int(32) entry_count;
//    for (i = 1 ; i <= entry_count ; i++){
//       SampleEntry(); // an instance of a class derived from SampleEntry
//    }
// }
type SampleDescriptionBox struct {
   BoxHeader  BoxHeader
   FullBoxHeader FullBoxHeader
   Entry_Count uint32
   Entries []*VisualSampleEntry
}

func (b *SampleDescriptionBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Entry_Count); err != nil {
      return err
   }
   b.Entries = make([]*VisualSampleEntry, b.Entry_Count)
   for i := range b.Entries {
      var entry VisualSampleEntry
      err := entry.Decode(r)
      if err != nil {
         return err
      }
      b.Entries[i] = &entry
   }
   return nil
}

func (s SampleDescriptionBox) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := s.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.Entry_Count); err != nil {
      return err
   }
   for _, entry := range s.Entries {
      err := entry.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
