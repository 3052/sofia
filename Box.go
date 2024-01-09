package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "io"
)

// 4.2.2 Object definitions
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

// 4.2.2 Object definitions
//  aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//     unsigned int(8) version = v;
//     bit(24) flags = f;
//  }
type FullBoxHeader struct {
   Version  uint8
   RawFlags [3]byte
}

func (f *FullBoxHeader) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, f)
}

func (f FullBoxHeader) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, f)
}

func (f FullBoxHeader) Flags() uint32 {
   var v uint32
   v |= uint32(f.RawFlags[0])<<16
   v |= uint32(f.RawFlags[1])<<8
   v |= uint32(f.RawFlags[2])
   return v
}

// unsigned int(32) type = boxtype;
func (b BoxHeader) BoxType() string {
   return string(b.Type[:])
}

// unsigned int(8)[16] usertype = extended_type;
func (b BoxHeader) Extended_Type() string {
   return hex.EncodeToString(b.UserType[:])
}

func (b *BoxHeader) Decode(r io.Reader) error {
   if err := binary.Read(r, binary.BigEndian, &b.BoxSize); err != nil {
      return err
   }
   if _, err := io.ReadFull(r, b.Type[:]); err != nil {
      return err
   }
   if b.BoxType() == "uuid" {
      _, err := io.ReadFull(r, b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (b BoxHeader) Encode(w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, b.BoxSize)
   if err != nil {
      return err
   }
   if _, err := w.Write(b.Type[:]); err != nil {
      return err
   }
   if b.BoxType() == "uuid" {
      _, err := w.Write(b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

// 4.2.2 Object definitions
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
   BoxSize uint32
   Type [4]uint8
   UserType [16]uint8
}

func (b BoxHeader) Size() uint32 {
   var s uint32 = 4 // BoxSize
   s += 4 // Type
   if b.BoxType() == "uuid" {
      s += 16 // UserType
   }
   return s
}

func (b BoxHeader) Reader(r io.Reader) io.Reader {
   n := int64(b.BoxSize - b.Size())
   return io.LimitReader(r, n)
}
