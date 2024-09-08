package sofia

import (
	"154.pages.dev/sofia/box"
	"io"
)

// ISO/IEC 14496-12
//
//	aligned(8) class SchemeInformationBox extends Box('schi') {
//	   Box scheme_specific_data[];
//	}
type SchemeInformation struct {
	BoxHeader       box.Header
	TrackEncryption TrackEncryption
}

func (s *SchemeInformation) read(r io.Reader) error {
	return s.TrackEncryption.read(r)
}

func (s SchemeInformation) write(w io.Writer) error {
	err := s.BoxHeader.Write(w)
	if err != nil {
		return err
	}
	return s.TrackEncryption.write(w)
}
