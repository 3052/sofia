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
   Size Size
   Type Type
}

func (b *BoxHeader) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, b)
}

// aligned(8) class FullBoxHeader(
//    unsigned int(8) v,
//    bit(24) f
// ) {
//    unsigned int(8) version = v;
//    bit(24) flags = f;
// }
type FullBoxHeader [4]byte

func (f FullBoxHeader) Decode(r io.Reader) error {
   _, err := r.Read(f[:])
   if err != nil {
      return err
   }
   return nil
}

type Size uint32

// aligned(8) class Box (
//    unsigned int(32) boxtype,
//    optional unsigned int(8)[16] extended_type
// ) {
//    BoxHeader(
//       boxtype,
//       extended_type
//    );
//    // the remaining bytes are the BoxPayload
// }
func (s Size) Payload() int64 {
   return int64(s) - 8
}

type Type [4]byte

func (t Type) String() string {
   return string(t[:])
}
