package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "io"
   "strconv"
)

// ISO/IEC 14496-12
//   aligned(8) class Box (
//      unsigned int(32) boxtype,
//      optional unsigned int(8)[16] extended_type
//   ) {
//      BoxHeader(boxtype, extended_type);
//      // the remaining bytes are the BoxPayload
//   }
type Box struct {
   BoxHeader BoxHeader
   Payload   []byte
}

func (b *Box) read(r io.Reader) error {
   _, size := b.BoxHeader.get_size()
   b.Payload = make([]byte, size)
   _, err := io.ReadFull(r, b.Payload)
   if err != nil {
      return err
   }
   return nil
}

func (b *Box) write(w io.Writer) error {
   err := b.BoxHeader.write(w)
   if err != nil {
      return err
   }
   _, err = w.Write(b.Payload)
   if err != nil {
      return err
   }
   return nil
}

// ISO/IEC 14496-12
//   aligned(8) class BoxHeader (
//      unsigned int(32) boxtype,
//      optional unsigned int(8)[16] extended_type
//   ) {
//      unsigned int(32) size;
//      unsigned int(32) type = boxtype;
//      if (size==1) {
//         unsigned int(64) largesize;
//      } else if (size==0) {
//         // box extends to end of file
//      }
//      if (boxtype=='uuid') {
//         unsigned int(8)[16] usertype = extended_type;
//      }
//   }
type BoxHeader struct {
   Size     uint32
   Type     Type
   UserType UUID
}

func (b *BoxHeader) get_size() (int, int64) {
   size := binary.Size(b.Size)
   size += binary.Size(b.Type)
   if b.Type.String() == "uuid" {
      size += binary.Size(b.UserType)
   }
   return size, int64(b.Size) - int64(size)
}

func (b *BoxHeader) Read(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.Size)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(r, b.Type[:])
   if err != nil {
      return err
   }
   if b.Type.String() == "uuid" {
      _, err := io.ReadFull(r, b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (b *BoxHeader) write(w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, b.Size)
   if err != nil {
      return err
   }
   _, err = w.Write(b.Type[:])
   if err != nil {
      return err
   }
   if b.Type.String() == "uuid" {
      _, err := w.Write(b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

// ISO/IEC 14496-12
//   aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//      unsigned int(8) version = v;
//      bit(24) flags = f;
//   }
type FullBoxHeader struct {
   Version uint8
   Flags   [3]byte
}

func (f *FullBoxHeader) get_flags() uint32 {
   var flag [4]byte
   copy(flag[1:], f.Flags[:])
   return binary.BigEndian.Uint32(flag[:])
}

func (f *FullBoxHeader) read(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, f)
}

func (f FullBoxHeader) write(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, f)
}

type UUID [16]uint8

func (u UUID) String() string {
   return hex.EncodeToString(u[:])
}

func (t Type) String() string {
   return string(t[:])
}

type Type [4]uint8

type box_error struct {
   container Type
   box_type Type
}

func (b box_error) Error() string {
   c := []byte("container:")
   c = strconv.AppendQuote(c, b.container.String())
   c = append(c, " box type:"...)
   c = strconv.AppendQuote(c, b.box_type.String())
   return string(c)
}
