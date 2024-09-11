package tenc

import (
   "154.pages.dev/sofia"
   "encoding/binary"
   "io"
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

func (b *Box) Read(src io.Reader) error {
   err := b.BoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.Extends)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(src, b.DefaultKid[:])
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
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.Extends)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.DefaultKid[:])
   if err != nil {
      return err
   }
   return nil
}
