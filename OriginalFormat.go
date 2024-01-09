package sofia

import "io"

// Container: ProtectionSchemeInfoBox
//  aligned(8) class OriginalFormatBox(codingname) extends Box('frma') {
//     unsigned int(32) data_format = codingname;
//     // format of decrypted, encoded data (in case of protection)
//     // or un-transformed sample entry (in case of restriction
//     // and complete track information)
//  }
type OriginalFormatBox struct {
   BoxHeader BoxHeader
   DataFormat [4]uint8
}

func (b *OriginalFormatBox) Decode(r io.Reader) error {
   _, err := io.ReadFull(r, b.DataFormat[:])
   if err != nil {
      return err
   }
   return nil
}

func (b OriginalFormatBox) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(b.DataFormat[:]); err != nil {
      return err
   }
   return nil
}
