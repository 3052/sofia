package sofia

import (
   "encoding/binary"
   "io"
)

func (b BoxHeader) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, b)
}

func (b BoxHeader) BoxPayload() int64 {
   return int64(b.Size) - 8
}

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
   RawType [4]byte
}

func (b BoxHeader) Type() string {
   return string(b.RawType[:])
}

func (b *BoxHeader) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, b)
}

func (f *FullBoxHeader) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, f)
}

func (f FullBoxHeader) Flags() uint32 {
   b := []byte{0}
   b = append(b, f.RawFlags[:]...)
   return binary.BigEndian.Uint32(b)
}

// aligned(8) class FullBoxHeader(
//    unsigned int(8) v,
//    bit(24) f
// ) {
//    unsigned int(8) version = v;
//    bit(24) flags = f;
// }
type FullBoxHeader struct {
   Version uint8
   RawFlags [3]byte
}

func (f FullBoxHeader) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, f)
}

func (b Box) Encode(w io.Writer) error {
   err := b.Header.Encode(w)
   if err != nil {
      return err
   }
   _, err = w.Write(b.Payload)
   if err != nil {
      return err
   }
   return nil
}

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
type Box struct {
   Header BoxHeader
   Payload []byte
}
