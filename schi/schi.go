package schi

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/tenc"
   "io"
)

// ISO/IEC 14496-12
//
//   aligned(8) class SchemeInformationBox extends Box('schi') {
//      Box scheme_specific_data[];
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Tenc      tenc.Box
}

func (b *Box) Read(src io.Reader) error {
   return b.Tenc.Read(src)
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   return b.Tenc.Write(dst)
}
