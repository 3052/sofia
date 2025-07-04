package pssh

import (
   "41.neocities.org/sofia"
   "encoding/binary"
)

// ISO/IEC 23001-7
//   aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
//      'pssh', version, flags=0,
//   ) {
//      unsigned int(8)[16] SystemID;
//      if (version > 0) {
//         unsigned int(32) KID_count;
//         {
//            unsigned int(8)[16] KID;
//         } [KID_count];
//      }
//      unsigned int(32) DataSize;
//      unsigned int(8)[DataSize] Data;
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   SystemId      sofia.Uuid
   KidCount      uint32
   Kid           []sofia.Uuid
   DataSize      uint32
   Data          []uint8
}

func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
   if err != nil {
      return err
   }
   data = data[n:]
   b.SystemId = [16]byte(data)
   data = data[16:]
   if b.FullBoxHeader.Version > 0 {
      n, err := binary.Decode(data, binary.BigEndian, &b.KidCount)
      if err != nil {
         return err
      }
      data = data[n:]
      b.Kid = make([]sofia.Uuid, b.KidCount)
      n, err = binary.Decode(data, binary.BigEndian, b.Kid)
      if err != nil {
         return err
      }
      data = data[n:]
   }
   n, err = binary.Decode(data, binary.BigEndian, &b.DataSize)
   if err != nil {
      return err
   }
   data = data[n:]
   b.Data = data[:b.DataSize]
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
   data = append(data, b.SystemId[:]...)
   if b.FullBoxHeader.Version > 0 {
      data = binary.BigEndian.AppendUint32(data, b.KidCount)
      data, err = binary.Append(data, binary.BigEndian, b.Kid)
      if err != nil {
         return nil, err
      }
   }
   data = binary.BigEndian.AppendUint32(data, b.DataSize)
   return append(data, b.Data...), nil
}

