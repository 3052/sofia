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

// dashif.org/identifiers/content_protection
func (b *Box) Widevine() bool {
   return b.SystemId.String() == "edef8ba979d64acea3c827dcd51d21ed"
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
   buf = append(buf, b.SystemId[:]...)
   if b.FullBoxHeader.Version > 0 {
      buf = binary.BigEndian.AppendUint32(buf, b.KidCount)
      buf, err = binary.Append(buf, binary.BigEndian, b.Kid)
      if err != nil {
         return nil, err
      }
   }
   buf = binary.BigEndian.AppendUint32(buf, b.DataSize)
   return append(buf, b.Data...), nil
}

func (b *Box) Read(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n = copy(b.SystemId[:], buf)
   buf = buf[n:]
   if b.FullBoxHeader.Version > 0 {
      n, err := binary.Decode(buf, binary.BigEndian, &b.KidCount)
      if err != nil {
         return err
      }
      buf = buf[n:]
      b.Kid = make([]sofia.Uuid, b.KidCount)
      n, err = binary.Decode(buf, binary.BigEndian, b.Kid)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   n, err = binary.Decode(buf, binary.BigEndian, &b.DataSize)
   if err != nil {
      return err
   }
   buf = buf[n:]
   b.Data = buf[:b.DataSize]
   return nil
}
