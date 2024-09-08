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

func (o *Box) read(r io.Reader) error {
   _, err := io.ReadFull(r, o.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}

func (o Box) write(w io.Writer) error {
   err := o.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   _, err = w.Write(o.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}
