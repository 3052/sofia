package tenc

import (
   "154.pages.dev/sofia"
   "encoding/binary"
)

// ISO/IEC 23001-7
//   aligned(8) class TrackEncryptionBox extends FullBox('tenc', version, flags=0) {
//      unsigned int(8) reserved = 0;
//      if (version==0) {
//         unsigned int(8) reserved = 0;
//      } else { // version is 1 or greater
//         unsigned int(4) default_crypt_byte_block;
//         unsigned int(4) default_skip_byte_block;
//      }
//      unsigned int(8) default_isProtected;
//      unsigned int(8) default_Per_Sample_IV_Size;
//      unsigned int(8)[16] default_KID;
//      if (default_isProtected ==1 && default_Per_Sample_IV_Size == 0) {
//         unsigned int(8) default_constant_IV_size;
//         unsigned int(8)[default_constant_IV_size] default_constant_IV;
//      }
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   Extends       struct {
      _                      uint8
      _                      uint8
      DefaultIsProtected     uint8
      DefaultPerSampleIvSize uint8
   }
   DefaultKid sofia.Uuid
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.Extends)
   if err != nil {
      return nil, err
   }
   return append(buf, b.DefaultKid[:]...), nil
}

func (b *Box) Decode(buf []byte) error {
   ns, err := b.BoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   n, err := b.FullBoxHeader.Decode(buf[ns:])
   if err != nil {
      return err
   }
   ns += n
   n, err = binary.Decode(buf[ns:], binary.BigEndian, &b.Extends)
   if err != nil {
      return err
   }
   ns += n
   copy(b.DefaultKid[:], buf[ns:])
   return nil
}
