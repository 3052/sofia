package sofia

import (
	"154.pages.dev/sofia/box"
	"io"
)

// ISO/IEC 14496-12
//
//	aligned(8) class SampleTableBox extends Box('stbl') {
//	}
type SampleTable struct {
	BoxHeader         box.Header
	Boxes             []box.Box
	SampleDescription SampleDescription
}

func (s *SampleTable) read(r io.Reader, size int64) error {
	r = io.LimitReader(r, size)
	for {
		var head box.Header
		err := head.Read(r)
		switch err {
		case nil:
			switch head.Type.String() {
			case "stsd":
				_, size := head.GetSize()
				s.SampleDescription.BoxHeader = head
				err := s.SampleDescription.read(r, size)
				if err != nil {
					return err
				}
			case "sgpd", // Paramount
				"stco", // Roku
				"stsc", // Roku
				"stss", // CineMember
				"stsz", // Roku
				"stts": // Roku
				object := box.Box{BoxHeader: head}
				err := object.Read(r)
				if err != nil {
					return err
				}
				s.Boxes = append(s.Boxes, object)
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

func (s SampleTable) write(w io.Writer) error {
	err := s.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	for _, object := range s.Boxes {
		err := object.Write(w)
		if err != nil {
			return err
		}
	}
	return s.SampleDescription.write(w)
}
