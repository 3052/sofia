package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "io"
)

func (b *Box) Decode(r io.Reader) error {
   var err error
   b.Payload, err = io.ReadAll(r)
   if err != nil {
      return err
   }
   return nil
}

func (b Box) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(b.Payload); err != nil {
      return err
   }
   return nil
}

func (b *BoxHeader) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.S)
   if err != nil {
      return err
   }
   if b.GetType() == "uuid" {
      _, err := io.ReadFull(r, b.Usertype[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (b BoxHeader) Encode(w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, b.S)
   if err != nil {
      return err
   }
   if b.GetType() == "uuid" {
      _, err := w.Write(b.Usertype[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (f *FullBoxHeader) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, f)
}

func (f FullBoxHeader) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, f)
}

// ISO/IEC 14496-12
//  aligned(8) class Box (
//     unsigned int(32) boxtype,
//     optional unsigned int(8)[16] extended_type
//  ) {
//     BoxHeader(boxtype, extended_type);
//     // the remaining bytes are the BoxPayload
//  }
type Box struct {
   BoxHeader BoxHeader
   Payload []byte
}

func (b BoxHeader) GetUsertype() string {
   return hex.EncodeToString(b.Usertype[:])
}

func (f FullBoxHeader) GetFlags() uint32 {
   var b [4]byte
   copy(b[1:], f.Flags[:])
   return binary.BigEndian.Uint32(b[:])
}

// ISO/IEC 14496-12
//  aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//     unsigned int(8) version = v;
//     bit(24) flags = f;
//  }
type FullBoxHeader struct {
   Version  uint8
   Flags [3]byte
}

// ISO/IEC 14496-12
//  aligned(8) class BoxHeader (
//     unsigned int(32) boxtype,
//     optional unsigned int(8)[16] extended_type
//  ) {
//     unsigned int(32) size;
//     unsigned int(32) type = boxtype;
//     if (size==1) {
//        unsigned int(64) largesize;
//     } else if (size==0) {
//        // box extends to end of file
//     }
//     if (boxtype=='uuid') {
//        unsigned int(8)[16] usertype = extended_type;
//     }
//  }
type BoxHeader struct {
   S struct {
      Size uint32
      Type [4]uint8
   }
   Usertype [16]uint8
}

func (b BoxHeader) GetType() string {
   return string(b.S.Type[:])
}

func (b BoxHeader) Payload(r io.Reader) io.Reader {
   n := int64(b.S.Size - b.Size())
   return io.LimitReader(r, n)
}

func (FullBoxHeader) Size() uint32 {
   var s uint32 = 1 // Version
   return s + 3 // Flags
}

func (b BoxHeader) Size() uint32 {
   var s uint32 = 4 // size
   s += 4 // type
   if b.GetType() == "uuid" {
      s += 16 // usertype
   }
   return s
}
