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

type Reference [3]uint32

func (Reference) Size() uint32 {
   return 3 * 4
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) ReferencedSize() uint32 {
   return r[0] & sofia.Reference(r).Mask()
}

func (r Reference) SetReferencedSize(v uint32) {
   r[0] &= ^sofia.Reference(r).Mask()
   r[0] |= v
}

// 1. set timescale?
// 2. set references
// 3. set size
func (s *SegmentIndexBox) Global() {
   s.BoxHeader.BoxSize = s.Size()
   copy(s.BoxHeader.Type[:], "sidx")
   s.ReferenceId = 1
   s.ReferenceCount = uint16(len(s.Reference))
   // Reference []Reference
}
