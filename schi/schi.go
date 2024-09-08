package schi

import (
   "154.pages.dev/sofia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class SchemeInformationBox extends Box('schi') {
//      Box scheme_specific_data[];
//   }
type Box struct {
   BoxHeader       sofia.BoxHeader
   TrackEncryption TrackEncryption
}

func (s *Box) read(r io.Reader) error {
   return s.TrackEncryption.read(r)
}

func (s Box) write(w io.Writer) error {
   err := s.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   return s.TrackEncryption.write(w)
}
