package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class BoxHeader (
//    unsigned int(32) boxtype,
//    optional unsigned int(8)[16] extended_type
// ) {
//    unsigned int(32) size;
//    unsigned int(32) type = boxtype;
//    if (size==1) {
//       unsigned int(64) largesize;
//    } else if (size==0) {
//       // box extends to end of file
//    }
//    if (boxtype=='uuid') {
//       unsigned int(8)[16] usertype = extended_type;
//    }
// }
type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func (b *BoxHeader) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.Size)
   if err != nil {
      return err
   }
   _, err = r.Read(b.Type[:])
   if err != nil {
      return err
   }
   return nil
}

func (b BoxHeader) String() string {
   return string(b.Type[:])
}

// aligned(8) class FullBox(
//    unsigned int(32) boxtype,
//    unsigned int(8) v, bit(24) f,
//    optional unsigned int(8)[16] extended_type
// ) extends Box(boxtype, extended_type) {
//    FullBoxHeader(v, f);
//    // the remaining bytes are the FullBoxPayload
// }
type FullBox struct {
   BoxHeader BoxHeader
   Header FullBoxHeader
   Payload []byte
}

// aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//    unsigned int(8) version = v;
//    bit(24) flags = f;
// }
type FullBoxHeader struct {
   Version uint8
   Flags [3]byte
}

func (f *FullBoxHeader) Decode(r io.Reader) error {
   err := binary.Read(r, nil, &f.Version)
   if err != nil {
      return err
   }
   _, err = r.Read(f.Flags[:])
   if err != nil {
      return err
   }
   return nil
}

func (FullBoxHeader) Size() uint32 {
   return 4
}
