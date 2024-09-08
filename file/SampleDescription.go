package sofia

import (
	"154.pages.dev/sofia/box"
	"encoding/binary"
	"io"
)

func (s *SampleDescription) read(r io.Reader, size int64) error {
	r = io.LimitReader(r, size)
	err := s.FullBoxHeader.Read(r)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &s.EntryCount)
	if err != nil {
		return err
	}
	for {
		var head box.Header
		err := head.Read(r)
		switch err {
		case nil:
			_, size := head.GetSize()
			switch head.Type.String() {
			case "avc1", // Tubi
				"ec-3", // Max
				"mp4a": // Tubi
				value := box.Box{BoxHeader: head}
				err := value.Read(r)
				if err != nil {
					return err
				}
				s.Boxes = append(s.Boxes, value)
			case "enca":
				s.AudioSample = &AudioSampleEntry{}
				s.AudioSample.SampleEntry.BoxHeader = head
				err := s.AudioSample.read(r, size)
				if err != nil {
					return err
				}
			case "encv":
				s.VisualSample = &VisualSampleEntry{}
				s.VisualSample.SampleEntry.BoxHeader = head
				err := s.VisualSample.read(r, size)
				if err != nil {
					return err
				}
			default:
				return box.Error{s.BoxHeader.Type, head.Type}
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

// ISO/IEC 14496-12
//
//	aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//	   int i ;
//	   unsigned int(32) entry_count;
//	   for (i = 1 ; i <= entry_count ; i++){
//	      SampleEntry(); // an instance of a class derived from SampleEntry
//	   }
//	}
type SampleDescription struct {
	BoxHeader     box.Header
	FullBoxHeader box.FullBoxHeader
	EntryCount    uint32
	Boxes         []box.Box
	AudioSample   *AudioSampleEntry
	VisualSample  *VisualSampleEntry
}

func (s SampleDescription) Protection() (*ProtectionSchemeInfo, bool) {
	if v := s.AudioSample; v != nil {
		return &v.ProtectionScheme, true
	}
	if v := s.VisualSample; v != nil {
		return &v.ProtectionScheme, true
	}
	return nil, false
}

func (s SampleDescription) SampleEntry() (*SampleEntry, bool) {
	if v := s.AudioSample; v != nil {
		return &v.SampleEntry, true
	}
	if v := s.VisualSample; v != nil {
		return &v.SampleEntry, true
	}
	return nil, false
}

func (s SampleDescription) write(w io.Writer) error {
	err := s.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = s.FullBoxHeader.Write(w)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.BigEndian, s.EntryCount)
	if err != nil {
		return err
	}
	for _, value := range s.Boxes {
		err := value.Write(w)
		if err != nil {
			return err
		}
	}
	if s.AudioSample != nil {
		err := s.AudioSample.write(w)
		if err != nil {
			return err
		}
	}
	if s.VisualSample != nil {
		err := s.VisualSample.write(w)
		if err != nil {
			return err
		}
	}
	return nil
}
