package tenc

import "41.neocities.org/sofia"

func (b *Box) Read(data []byte) error {
   n, err := b.BoxHeader.Decode(data)
   if err != nil {
      return err
   }
   data = data[n:]
   b.FullBoxHeader.Version = data[0]
   data = data[1:]
   n = copy(b.FullBoxHeader.Flags[:], data)
   data = data[n:]
   data = data[1:] // reserved
   b.ByteBlock = data[0]
   data = data[1:]
   b.DefaultIsProtected = data[0]
   data = data[1:]
   b.DefaultPerSampleIvSize = data[0]
   data = data[1:]
   n = copy(b.DefaultKid[:], data)
   data = data[n:]
   if b.DefaultPerSampleIvSize == 0 {
      if b.DefaultIsProtected == 1 {
         b.DefaultConstantIvSize, data = data[0], data[1:]
         b.DefaultConstantIv = data[:b.DefaultConstantIvSize]
      }
   }
   return nil
}

// ISO/IEC 23001-7
//
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
   BoxHeader              sofia.BoxHeader
   FullBoxHeader          sofia.FullBoxHeader
   _                      uint8
   ByteBlock              uint8
   DefaultIsProtected     uint8
   DefaultPerSampleIvSize uint8
   DefaultKid             sofia.Uuid
   DefaultConstantIvSize  uint8
   DefaultConstantIv      []uint8
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data = append(data, b.FullBoxHeader.Version)
   data = append(data, b.FullBoxHeader.Flags[:]...)
   data = append(data, 0, // reserved
      b.ByteBlock,
      b.DefaultIsProtected,
      b.DefaultPerSampleIvSize,
   )
   data = append(data, b.DefaultKid[:]...)
   if b.DefaultPerSampleIvSize == 0 {
      if b.DefaultIsProtected == 1 {
         data = append(data, b.DefaultConstantIvSize)
         data = append(data, b.DefaultConstantIv...)
      }
   }
   return data, nil
}
