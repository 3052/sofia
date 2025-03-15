package mdhd

import (
   "41.neocities.org/sofia"
   "encoding/binary"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MediaHeaderBox extends FullBox('mdhd', version, 0) {
//      if (version==1) {
//         unsigned int(64) creation_time;
//         unsigned int(64) modification_time;
//         unsigned int(32) timescale;
//         unsigned int(64) duration;
//      } else { // version==0
//         unsigned int(32) creation_time;
//         unsigned int(32) modification_time;
//         unsigned int(32) timescale;
//         unsigned int(32) duration;
//      }
//      bit(1) pad = 0;
//      unsigned int(5)[3] language; // ISO-639-2/T language code
//      unsigned int(16) pre_defined = 0;
//   }
type Box struct {
   BoxHeader        sofia.BoxHeader
   FullBoxHeader    sofia.FullBoxHeader
   
   CreationTime     []byte
   ModificationTime []byte
   Timescale        uint32
   Duration         []byte
   Language         uint16
   PreDefined       uint16
}

func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
   if err != nil {
      return err
   }
   data = data[n:]
   if b.FullBoxHeader.Version == 1 {
   } else {
   }
}

//func (b *Box) Append(data []byte) ([]byte, error) {
//   data, err := b.BoxHeader.Append(data)
//   if err != nil {
//      return nil, err
//   }
//   return append(data, b.DataFormat[:]...), nil
//}
