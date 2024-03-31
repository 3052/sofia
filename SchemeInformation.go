package sofia

import "io"

// ISO/IEC 14496-12
//  aligned(8) class SchemeInformationBox extends Box('schi') {
//     Box scheme_specific_data[];
//  }
type SchemeInformation struct {
   Box Box
}

func (s *SchemeInformation) read(r io.Reader) error {
   return s.Box.read(r)
}

func (s SchemeInformation) write(w io.Writer) error {
   return s.Box.write(w)
}
