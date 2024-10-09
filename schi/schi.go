package schi

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/tenc"
)

// ISO/IEC 14496-12
//   aligned(8) class SchemeInformationBox extends Box('schi') {
//      Box scheme_specific_data[];
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Tenc      tenc.Box
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   return b.Tenc.Append(buf)
}

func (b *Box) Read(buf []byte) error {
   return b.Tenc.Read(buf)
}
