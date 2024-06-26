package sofia

import "io"

// ISO/IEC 14496-12
//  aligned(8) class OriginalFormatBox(codingname) extends Box('frma') {
//     unsigned int(32) data_format = codingname;
//     // format of decrypted, encoded data (in case of protection)
//     // or un-transformed sample entry (in case of restriction
//     // and complete track information)
//  }
type OriginalFormat struct {
   BoxHeader  BoxHeader
   DataFormat [4]uint8
}

func (o *OriginalFormat) read(r io.Reader) error {
   _, err := io.ReadFull(r, o.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}

func (o OriginalFormat) write(w io.Writer) error {
   err := o.BoxHeader.write(w)
   if err != nil {
      return err
   }
   _, err = w.Write(o.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}
