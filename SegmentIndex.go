package sofia

import (
	"encoding/binary"
	"io"
	"strconv"
)

func (s *SegmentIndex) Append(size uint32) {
	var r Reference
	r.SetReferencedSize(size)
	s.Reference = append(s.Reference, r)
	s.ReferenceCount++
	s.BoxHeader.Size = uint32(s.get_size())
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) ReferencedSize() uint32 {
	return r[0] & r.mask()
}

type Range struct {
	Start uint64
	End   uint64
}

func (r Range) String() string {
	b := []byte("bytes=")
	b = strconv.AppendUint(b, r.Start, 10)
	b = append(b, '-')
	b = strconv.AppendUint(b, r.End, 10)
	return string(b)
}

func (r *Reference) Decode(src io.Reader) error {
	return binary.Read(src, binary.BigEndian, r)
}

func (r Reference) Encode(dst io.Writer) error {
	return binary.Write(dst, binary.BigEndian, r)
}

func (r Reference) SetReferencedSize(v uint32) {
	r[0] &= ^r.mask()
	r[0] |= v
}

func (Reference) mask() uint32 {
	return 0xFFFFFFFF >> 1
}

func (s *SegmentIndex) Decode(r io.Reader) error {
	if err := s.FullBoxHeader.read(r); err != nil {
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

func (s SegmentIndex) Encode(w io.Writer) error {
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

func (s *SegmentIndex) New() {
	copy(s.BoxHeader.Type[:], "sidx")
}

// size will always fit inside 31 bits:
// unsigned int(31) referenced_size
// but range-start and range-end can both exceed 32 bits, so we must use 64 bit
func (s SegmentIndex) Ranges(start uint64) []Range {
	ranges := make([]Range, s.ReferenceCount)
	for i, ref := range s.Reference {
		size := uint64(ref.ReferencedSize())
		ranges[i] = Range{start, start + size - 1}
		start += size
	}
	return ranges
}

// ISO/IEC 14496-12
//
//	aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
//	   unsigned int(32) reference_ID;
//	   unsigned int(32) timescale;
//	   if (version==0) {
//	      unsigned int(32) earliest_presentation_time;
//	      unsigned int(32) first_offset;
//	   } else {
//	      unsigned int(64) earliest_presentation_time;
//	      unsigned int(64) first_offset;
//	   }
//	   unsigned int(16) reserved = 0;
//	   unsigned int(16) reference_count;
//	   for(i=1; i <= reference_count; i++) {
//	      bit (1) reference_type;
//	      unsigned int(31) referenced_size;
//	      unsigned int(32) subsegment_duration;
//	      bit(1) starts_with_SAP;
//	      unsigned int(3) SAP_type;
//	      unsigned int(28) SAP_delta_time;
//	   }
//	}
type SegmentIndex struct {
	BoxHeader                BoxHeader
	FullBoxHeader            FullBoxHeader
	ReferenceId              uint32
	Timescale                uint32
	EarliestPresentationTime []byte
	FirstOffset              []byte
	Reserved                 uint16
	ReferenceCount           uint16
	Reference                []Reference
}

type Reference [3]uint32

func (s SegmentIndex) get_size() int {
	v := s.BoxHeader.get_size()
	v += binary.Size(s.FullBoxHeader)
	v += binary.Size(s.ReferenceId)
	v += binary.Size(s.Timescale)
	v += binary.Size(s.EarliestPresentationTime)
	v += binary.Size(s.FirstOffset)
	v += binary.Size(s.Reserved)
	v += binary.Size(s.ReferenceCount)
	return v + binary.Size(s.Reference)
}
