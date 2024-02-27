package sofia

import "154.pages.dev/sofia"

func (s SegmentIndexBox) Size() uint32 {
   v := s.BoxHeader.Size()
   v += s.FullBoxHeader.Size()
   v += 4 // reference_ID
   v += 4 // timescale
   if s.FullBoxHeader.Version == 0 {
      v += 4 // earliest_presentation_time
      v += 4 // first_offset
   } else {
      v += 8 // earliest_presentation_time
      v += 8 // first_offset
   }
   v += 2 // reserved
   v += 2 // reference_count
   for _, r := range s.Reference {
      v += r.Size()
   }
   return v
}

type Reference [3]uint32

func (Reference) Size() uint32 {
   return 3 * 4
}

func (r Reference) SetReferencedSize(v uint32) {
   r[0] &= ^sofia.Reference(r).Mask()
   r[0] |= v
}

type SegmentIndexBox struct {
   BoxHeader sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   ReferenceId uint32
   Timescale uint32
   EarliestPresentationTime []byte
   FirstOffset []byte
   Reserved uint16
   ReferenceCount uint16
   Reference []Reference
}

func (s *SegmentIndexBox) Global() {
   s.BoxHeader.BoxSize = s.Size()
   copy(s.BoxHeader.Type[:], "sidx")
   s.ReferenceId = 1
   s.ReferenceCount = uint16(len(s.Reference))
}
