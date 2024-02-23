package sofia

import (
   "encoding/binary"
   "io"
)

// return a slice so we can measure progress
func (s SegmentIndexBox) ByteRanges(start uint32) [][2]uint32 {
   ranges := make([][2]uint32, s.ReferenceCount)
   for i, ref := range s.Reference {
      size := ref.ReferencedSize()
      ranges[i] = [2]uint32{start, start + size - 1}
      start += size
   }
   return ranges
}

func (r *Reference) Decode(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, r)
}

func (r Reference) Encode(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, r)
}

func (s *SegmentIndexBox) Decode(r io.Reader) error {
   if err := s.FullBoxHeader.Decode(r); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.ReferenceId); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.Timescale); err != nil {
      return err
   }
   if s.FullBoxHeader.Version == 0 {
      s.EarliestPresentationTime = make([]byte, 4)
      s.FirstOffset = make([]byte, 4)
   } else {
      s.EarliestPresentationTime = make([]byte, 8)
      s.FirstOffset = make([]byte, 8)
   }
   if _, err := io.ReadFull(r, s.EarliestPresentationTime); err != nil {
      return err
   }
   if _, err := io.ReadFull(r, s.FirstOffset); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.Reserved); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.ReferenceCount); err != nil {
      return err
   }
   s.Reference = make([]Reference, s.ReferenceCount)
   for i, ref := range s.Reference {
      err := ref.Decode(r)
      if err != nil {
         return err
      }
      s.Reference[i] = ref
   }
   return nil
}

func (s SegmentIndexBox) Encode(w io.Writer) error {
   if err := s.BoxHeader.Encode(w); err != nil {
      return err
   }
   if err := s.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.ReferenceId); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.Timescale); err != nil {
      return err
   }
   if _, err := w.Write(s.EarliestPresentationTime); err != nil {
      return err
   }
   if _, err := w.Write(s.FirstOffset); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.Reserved); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.ReferenceCount); err != nil {
      return err
   }
   for _, ref := range s.Reference {
      err := ref.Encode(w)
      if err != nil {
         return err
      }
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
   ReferenceId uint32
   Timescale uint32
   EarliestPresentationTime []byte
   FirstOffset []byte
   Reserved uint16
   ReferenceCount uint16
   Reference []Reference
}

type Reference [3]uint32

func (Reference) Mask() uint32 {
   return 0xFFFFFFFF>>1
}

func (r Reference) ReferencedSize() uint32 {
   return r[0] & r.Mask()
}
