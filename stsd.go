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
   SampleEntry []SampleEntry
}

func (b *SampleDescriptionBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Entry_Count); err != nil {
      return err
   }
   return nil
}
