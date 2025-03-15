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
      n = 8
   } else {
      n = 4
   }
   b.CreationTime, data = data[:n], data[n:]
   b.ModificationTime, data = data[:n], data[n:]
   b.Timescale, data = binary.BigEndian.Uint32(data), data[4:]
   b.Duration, data = data[:n], data[n:]
   b.Language, data = binary.BigEndian.Uint16(data), data[2:]
   b.PreDefined = binary.BigEndian.Uint16(data)
   return nil
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, b.FullBoxHeader)
   if err != nil {
      return nil, err
   }
   data = append(data, b.CreationTime...)
   data = append(data, b.ModificationTime...)
   data = binary.BigEndian.AppendUint32(data, b.Timescale)
   data = append(data, b.Duration...)
   data = binary.BigEndian.AppendUint16(data, b.Language)
   return binary.BigEndian.AppendUint16(data, b.PreDefined), nil
}
