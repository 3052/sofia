package mdat

import (
   "io"
   "sofia/box"
)

// ISO/IEC 14496-12
//  aligned(8) class MediaDataBox extends Box('mdat') {
//     bit(8) data[];
//  }
type Box struct {
   Box box.Box
}

func (b *Box) Read(r io.Reader) error {
   return b.Box.Read(r)
}

func (b *Box) Write(w io.Writer) error {
   return b.Box.Write(w)
}
