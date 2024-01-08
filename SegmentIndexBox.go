package sofia

import (
   "encoding/binary"
   "io"
)

type Reference [3]uint32

func (r *Reference) Decode(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, r)
}

func (r Reference) Encode(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, r)
}

func (r Reference) Referenced_Size() uint32 {
   return r[0] & (0xFFFFFFFF>>1)
}

func (s SegmentIndexBox) ByteRanges(start uint32) [][2]uint32 {
   ranges := make([][2]uint32, s.B.Reference_Count)
   for i, ref := range s.References {
      size := ref.Referenced_Size()
      ranges[i] = [2]uint32{start, start + size - 1}
      start += size
   }
   return ranges
}

func (s *SegmentIndexBox) Decode(r io.Reader) error {
   if err := s.FullBoxHeader.Decode(r); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.A); err != nil {
      return err
   }
   if s.FullBoxHeader.Version == 0 {
      s.Earliest_Presentation_Time = make([]byte, 4)
      s.First_Offset = make([]byte, 4)
   } else {
      s.Earliest_Presentation_Time = make([]byte, 8)
      s.First_Offset = make([]byte, 8)
   }
   if _, err := io.ReadFull(r, s.Earliest_Presentation_Time); err != nil {
      return err
   }
   if _, err := io.ReadFull(r, s.First_Offset); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.B); err != nil {
      return err
   }
   s.References = make([]Reference, s.B.Reference_Count)
   for i, ref := range s.References {
      err := ref.Decode(r)
      if err != nil {
         return err
      }
      s.References[i] = ref
   }
   return nil
}

// Container: File
//  aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
//     unsigned int(32) reference_ID;
//     unsigned int(32) timescale;
//     if (version==0) {
//        unsigned int(32) earliest_presentation_time;
//        unsigned int(32) first_offset;
//     } else {
//        unsigned int(64) earliest_presentation_time;
//        unsigned int(64) first_offset;
//     }
//     unsigned int(16) reserved = 0;
//     unsigned int(16) reference_count;
//     for(i=1; i <= reference_count; i++) {
//        bit (1) reference_type;
//        unsigned int(31) referenced_size;
//        unsigned int(32) subsegment_duration;
//        bit(1) starts_with_SAP;
//        unsigned int(3) SAP_type;
//        unsigned int(28) SAP_delta_time;
//     }
//  }
type SegmentIndexBox struct {
   BoxHeader BoxHeader
   FullBoxHeader FullBoxHeader
   A struct {
      Reference_ID uint32
      Timescale uint32
   }
   Earliest_Presentation_Time []byte
   First_Offset []byte
   B struct {
      Reserved uint16
      Reference_Count uint16
   }
   References []Reference
}

func (s SegmentIndexBox) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := s.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.A); err != nil {
      return err
   }
   if _, err := w.Write(s.Earliest_Presentation_Time); err != nil {
      return err
   }
   if _, err := w.Write(s.First_Offset); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.B); err != nil {
      return err
   }
   for _, ref := range s.References {
      err := ref.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
