package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "io"
)

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
   Size uint32
   // Type is used outside this module, so we cannot wrap it with Size:
   Type     [4]uint8
   Usertype [16]uint8
}

func (b BoxHeader) get_size() (int, int64) {
   s := binary.Size(b.Size)
   s += binary.Size(b.Type)
   if b.GetType() == "uuid" {
      s += binary.Size(b.Usertype)
   }
   return s, int64(b.Size) - int64(s)
}

func (b BoxHeader) payload(r io.Reader) io.Reader {
   _, n := b.get_size()
   return io.LimitReader(r, int64(n))
}

// ISO/IEC 14496-12
//
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

func (f *FullBoxHeader) read(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, f)
}

// ISO/IEC 14496-12
//
//   aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//      unsigned int(8) version = v;
//      bit(24) flags = f;
//   }
type FullBoxHeader struct {
   Version uint8
   Flags   [3]byte
}

func (f FullBoxHeader) get_flags() uint32 {
   var b [4]byte
   copy(b[1:], f.Flags[:])
   return binary.BigEndian.Uint32(b[:])
}

func (b BoxHeader) GetType() string {
   return string(b.Type[:])
}

// ISO/IEC 14496-12
//
func (b BoxHeader) get_usertype() string {
   return hex.EncodeToString(b.Usertype[:])
}

func (f FullBoxHeader) write(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, f)
}

func (b BoxHeader) write(w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, b.Size)
   if err != nil {
      return err
   }
   if _, err := w.Write(b.Type[:]); err != nil {
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

func (b Box) write(w io.Writer) error {
   err := b.BoxHeader.write(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(b.Payload); err != nil {
      return err
   }
   return nil
}

func (b *BoxHeader) read(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.Size)
   if err != nil {
      return err
   }
   if _, err := io.ReadFull(r, b.Type[:]); err != nil {
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

func (b *Box) read(r io.Reader) error {
   _, size := b.BoxHeader.get_size()
   b.Payload = make([]byte, size)
   _, err := io.ReadFull(r, b.Payload)
   if err != nil {
      return err
   }
   return nil
}
