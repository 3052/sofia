package sofia

import (
   "encoding/binary"
   "io"
)

// Container: SampleTableBox
//  aligned(8) class SampleSizeBox extends FullBox('stsz', version = 0, 0) {
//     unsigned int(32) sample_size;
//     unsigned int(32) sample_count;
//     if (sample_size==0) {
//        for (i=1; i <= sample_count; i++) {
//           unsigned int(32) entry_size;
//        }
//     }
//  }
type SampleSizeBox struct {
   BoxHeader     BoxHeader
   FullBoxHeader FullBoxHeader
   Sample_Size uint32
   Sample_Count uint32
   Entry_Size []uint32
}

func (b *SampleSizeBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Sample_Size); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Sample_Count); err != nil {
      return err
   }
   b.Entry_Size = make([]uint32, b.Sample_Count)
   for i, size := range b.Entry_Size {
      err := binary.Read(r, binary.BigEndian, &size)
      if err != nil {
         return err
      }
      b.Entry_Size[i] = size
   }
   return nil
}

func (b SampleSizeBox) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := b.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, b.Sample_Size); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, b.Sample_Count); err != nil {
      return err
   }
   for _, size := range b.Entry_Size {
      err := binary.Write(w, binary.BigEndian, size)
      if err != nil {
         return err
      }
   }
   return nil
}
