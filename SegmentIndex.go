package sofia

import (
   "154.pages.dev/sofia/box"
   "encoding/binary"
   "io"
)

type Reference [3]uint32

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) ReferencedSize() uint32 {
   return r[0] & r.mask()
}

func (Reference) mask() uint32 {
   return 0xFFFFFFFF >> 1
}

func (r *Reference) read(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, r)
}

func (r Reference) set_referenced_size(v uint32) {
   r[0] &= ^r.mask()
   r[0] |= v
}

func (r Reference) write(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, r)
}

// ISO/IEC 14496-12
//
//   aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
//      unsigned int(32) reference_ID;
//      unsigned int(32) timescale;
//      if (version==0) {
//         unsigned int(32) earliest_presentation_time;
//         unsigned int(32) first_offset;
//      } else {
//         unsigned int(64) earliest_presentation_time;
//         unsigned int(64) first_offset;
//      }
//      unsigned int(16) reserved = 0;
//      unsigned int(16) reference_count;
//      for(i=1; i <= reference_count; i++) {
//         bit (1) reference_type;
//         unsigned int(31) referenced_size;
//         unsigned int(32) subsegment_duration;
//         bit(1) starts_with_SAP;
//         unsigned int(3) SAP_type;
//         unsigned int(28) SAP_delta_time;
//      }
//   }
type SegmentIndex struct {
   BoxHeader                box.Header
   FullBoxHeader            FullBoxHeader
   ReferenceId              uint32
   Timescale                uint32
   EarliestPresentationTime []byte
   FirstOffset              []byte
   Reserved                 uint16
   ReferenceCount           uint16
   Reference                []Reference
}

func (s *SegmentIndex) Append(size uint32) {
   var r Reference
   r.set_referenced_size(size)
   s.Reference = append(s.Reference, r)
   s.ReferenceCount++
   s.BoxHeader.Size = uint32(s.GetSize())
}

func (s *SegmentIndex) New() {
   copy(s.BoxHeader.Type[:], "sidx")
}

func (s SegmentIndex) GetSize() int {
   v, _ := s.BoxHeader.GetSize()
   v += binary.Size(s.FullBoxHeader)
   v += binary.Size(s.ReferenceId)
   v += binary.Size(s.Timescale)
   v += binary.Size(s.EarliestPresentationTime)
   v += binary.Size(s.FirstOffset)
   v += binary.Size(s.Reserved)
   v += binary.Size(s.ReferenceCount)
   return v + binary.Size(s.Reference)
}

func (s *SegmentIndex) read(r io.Reader) error {
   err := s.FullBoxHeader.read(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.ReferenceId)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.Timescale)
   if err != nil {
      return err
   }
   if s.FullBoxHeader.Version == 0 {
      s.EarliestPresentationTime = make([]byte, 4)
      s.FirstOffset = make([]byte, 4)
   } else {
      s.EarliestPresentationTime = make([]byte, 8)
      s.FirstOffset = make([]byte, 8)
   }
   _, err = io.ReadFull(r, s.EarliestPresentationTime)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(r, s.FirstOffset)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.Reserved)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.ReferenceCount)
   if err != nil {
      return err
   }
   s.Reference = make([]Reference, s.ReferenceCount)
   for i, value := range s.Reference {
      err := value.read(r)
      if err != nil {
         return err
      }
      s.Reference[i] = value
   }
   return nil
}

func (s SegmentIndex) write(w io.Writer) error {
   err := s.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   err = s.FullBoxHeader.write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.ReferenceId)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.Timescale)
   if err != nil {
      return err
   }
   _, err = w.Write(s.EarliestPresentationTime)
   if err != nil {
      return err
   }
   _, err = w.Write(s.FirstOffset)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.Reserved)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.ReferenceCount)
   if err != nil {
      return err
   }
   for _, value := range s.Reference {
      err := value.write(w)
      if err != nil {
         return err
      }
   }
   return nil
}
