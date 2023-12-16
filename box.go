package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class FullBox(
//    unsigned int(32) boxtype,
//    unsigned int(8) v, bit(24) f,
//    optional unsigned int(8)[16] extended_type
// ) extends Box(boxtype, extended_type) {
//    FullBoxHeader(v, f);
//    // the remaining bytes are the FullBoxPayload
// }
type FullBox struct {
   Box Box
   Header FullBoxHeader
   Payload []byte
}

func (f *FullBox) Decode(r io.Reader) error {
   err := f.Box.Decode(r)
   if err != nil {
      return err
   }
   err = f.Header.Decode(r)
   if err != nil {
      return err
   }
   f.Payload = make([]byte, f.Box.Header.Size)
   _, err = r.Read(f.Payload)
   if err != nil {
      return err
   }
   return nil
}

func (b *Box) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.Header.Size)
   if err != nil {
      return err
   }
   _, err = r.Read(b.Header.Type[:])
   if err != nil {
      return err
   }
   b.Payload = make([]byte, b.Header.Size)
   _, err = r.Read(b.Payload)
   if err != nil {
      return err
   }
   return nil
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
   err := binary.Read(r, nil, f.Version)
   if err != nil {
      return err
   }
   _, err = r.Read(f.Flags[:])
   if err != nil {
      return err
   }
   return nil
}

// aligned(8) class Box (
//    unsigned int(32) boxtype,
//    optional unsigned int(8)[16] extended_type
// ) {
//    BoxHeader(boxtype, extended_type);
//    // the remaining bytes are the BoxPayload
// }
type Box struct {
   Header BoxHeader
   Payload []byte
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
   Type [4]byte
}
