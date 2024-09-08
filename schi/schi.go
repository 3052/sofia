package schi

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/tenc"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class SchemeInformationBox extends Box('schi') {
//      Box scheme_specific_data[];
//   }
type Box struct {
   BoxHeader       sofia.BoxHeader
   TrackEncryption tenc.Box
}

func (b *Box) Read(r io.Reader) error {
   return b.TrackEncryption.Read(r)
}

func (b Box) Write(w io.Writer) error {
   err := b.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   return b.TrackEncryption.Write(w)
}