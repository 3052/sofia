package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "io"
   "strconv"
)

func (b *BoxHeader) Read(src io.Reader) error {
   err := binary.Read(src, binary.BigEndian, &b.Size)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(src, b.Type[:])
   if err != nil {
      return err
   }
   if b.Type.String() == "uuid" {
      _, err := io.ReadFull(src, b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

type Error struct {
   Container Type
   Box       Type
}

func (e Error) Error() string {
   c := []byte("container:")
   c = strconv.AppendQuote(c, e.Container.String())
   c = append(c, " box type:"...)
   c = strconv.AppendQuote(c, e.Box.String())
   return string(c)
}

func (f *FullBoxHeader) GetFlags() uint32 {
   var flag [4]byte
   copy(flag[1:], f.Flags[:])
   return binary.BigEndian.Uint32(flag[:])
}

func (f *FullBoxHeader) Read(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, f)
}

func (f *FullBoxHeader) Write(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, f)
}

func (s *SampleEntry) Read(src io.Reader) error {
   _, err := io.ReadFull(src, s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Read(src, binary.BigEndian, &s.DataReferenceIndex)
}

func (s *SampleEntry) Write(dst io.Writer) error {
   err := s.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   _, err = dst.Write(s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Write(dst, binary.BigEndian, s.DataReferenceIndex)
}

type Type [4]uint8

func (t Type) String() string {
   return string(t[:])
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

// ISO/IEC 14496-12
//   aligned(8) abstract class SampleEntry(
//      unsigned int(32) format
//   ) extends Box(format) {
//      const unsigned int(8)[6] reserved = 0;
//      unsigned int(16) data_reference_index;
//   }
type SampleEntry struct {
   BoxHeader          BoxHeader
   Reserved           [6]uint8
   DataReferenceIndex uint16
}

func (u Uuid) String() string {
   return hex.EncodeToString(u[:])
}

type Uuid [16]uint8

func (b *BoxHeader) Write(dst io.Writer) error {
   err := binary.Write(dst, binary.BigEndian, b.Size)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.Type[:])
   if err != nil {
      return err
   }
   if b.Type.String() == "uuid" {
      _, err := dst.Write(b.UserType[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.Payload)
   if err != nil {
      return err
   }
   return nil
}

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

func (b *Box) Read(src io.Reader) error {
   b.Payload = make([]byte, b.BoxHeader.PayloadSize())
   _, err := io.ReadFull(src, b.Payload)
   if err != nil {
      return err
   }
   return nil
}

func (b *BoxHeader) HeaderSize() int {
   size := binary.Size(b.Size)
   size += binary.Size(b.Type)
   if b.Type.String() == "uuid" {
      size += binary.Size(b.UserType)
   }
   return size
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
   UserType Uuid
}

func (b *BoxHeader) PayloadSize() int64 {
   return int64(b.Size) - int64(b.HeaderSize())
}
