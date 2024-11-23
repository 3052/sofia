package frma

import "41.neocities.org/sofia"

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

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   return append(data, b.DataFormat[:]...), nil
}

func (b *Box) Read(data []byte) error {
   copy(b.DataFormat[:], data)
   return nil
}
