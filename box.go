package sofia

import (
   "encoding/binary"
   "io"
)

func (s Size) Payload() int64 {
   return int64(s) - 8
}

// aligned(8) class Box (
//    unsigned int(32) boxtype,
//    optional unsigned int(8)[16] extended_type
// ) {
//    BoxHeader(boxtype, extended_type);
//    // the remaining bytes are the BoxPayload
// }

type Size uint32

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
   Size Size
   Type Type
}

type Type [4]byte

func (t Type) String() string {
   return string(t[:])
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
