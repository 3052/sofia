package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
//    'pssh',
//    version,
//    flags=0
// ) {
//    unsigned int(8)[16] SystemID;
//    if (version > 0) {
//       unsigned int(32) KID_count;
//       {
//          unsigned int(8)[16] KID;
//       } [KID_count];
//    }
//    unsigned int(32) DataSize;
//    unsigned int(8)[DataSize] Data;
// }
type ProtectionSystemSpecificHeader struct {
   FullBoxHeader FullBoxHeader
   SystemID [16]uint8
   DataSize uint32
   Data []uint8
}

func (p *ProtectionSystemSpecificHeader) Decode(r io.Reader) error {
   err := p.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   _, err = r.Read(p.SystemID[:])
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &p.DataSize)
   if err != nil {
      return err
   }
   p.Data = make([]uint8, p.DataSize)
   _, err = r.Read(p.Data)
   if err != nil {
      return err
   }
   return nil
}
