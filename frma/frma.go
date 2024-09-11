package frma

import (
   "154.pages.dev/sofia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class OriginalFormatBox(codingname) extends Box('frma') {
//      unsigned int(32) data_format = codingname;
//      // format of decrypted, encoded data (in case of protection)
//      // or un-transformed sample entry (in case of restriction
//      // and complete track information)
//   }
type Box struct {
   BoxHeader  sofia.BoxHeader
   DataFormat [4]uint8
}

func (b *Box) Read(src io.Reader) error {
   _, err := io.ReadFull(src, b.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}
