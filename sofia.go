package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "strconv"
)

// ISO/IEC 14496-12
//   aligned(8) class FullBoxHeader(unsigned int(8) v, bit(24) f) {
//      unsigned int(8) version = v;
//      bit(24) flags = f;
//   }
type FullBoxHeader struct {
   Version uint8
   Flags   [3]byte
}

func (b *BoxHeader) GetSize() int {
   size := binary.Size(b.Size)
   size += binary.Size(b.Type)
   if b.UserType != nil {
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
   UserType *Uuid
}

func (b *BoxHeader) Append(buf []byte) ([]byte, error) {
   buf = binary.BigEndian.AppendUint32(buf, b.Size)
   buf = append(buf, b.Type[:]...)
   if b.UserType != nil {
      buf = append(buf, (*b.UserType)[:]...)
   }
   return buf, nil
}

type SizeGetter interface {
   GetSize() int
}

type Appender interface {
   Append([]byte) ([]byte, error)
}

type Decoder interface {
   Decode([]byte) (int, error)
}

type Reader interface {
   Read([]byte) error
}

func (t Type) String() string {
   return string(t[:])
}

func (b *Box) Read(buf []byte) error {
   n, err := b.BoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   b.Payload = buf[n:b.BoxHeader.Size]
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

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   return append(buf, b.Payload...), nil
}

type Error struct {
   Container BoxHeader
   Box BoxHeader
}

func (e *Error) Error() string {
   buf := []byte("container:")
   buf = strconv.AppendQuote(buf, e.Container.Type.String())
   buf = append(buf, " box type:"...)
   buf = strconv.AppendQuote(buf, e.Box.Type.String())
   return string(buf)
}

func (f *FullBoxHeader) GetFlags() uint32 {
   var flag [4]byte
   copy(flag[1:], f.Flags[:])
   return binary.BigEndian.Uint32(flag[:])
}

func (f *FullBoxHeader) Append(buf []byte) ([]byte, error) {
   return binary.Append(buf, binary.BigEndian, f)
}

func (f *FullBoxHeader) Decode(buf []byte) (int, error) {
   return binary.Decode(buf, binary.BigEndian, f)
}

func (s *SampleEntry) Decode(buf []byte) (int, error) {
   ns := copy(s.Reserved[:], buf)
   n, err := binary.Decode(buf[ns:], binary.BigEndian, &s.DataReferenceIndex)
   if err != nil {
      return 0, err
   }
   return ns+n, nil
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

func (s *SampleEntry) Append(buf []byte) ([]byte, error) {
   buf, err := s.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf = append(buf, s.Reserved[:]...)
   return binary.BigEndian.AppendUint16(buf, s.DataReferenceIndex), nil
}

func (u Uuid) String() string {
   return hex.EncodeToString(u[:])
}

type Uuid [16]uint8

type Type [4]uint8

func (b *BoxHeader) Decode(buf []byte) (int, error) {
   n, err := binary.Decode(buf, binary.BigEndian, &b.Size)
   if err != nil {
      return 0, err
   }
   n += copy(b.Type[:], buf[n:])
   if b.Type.String() == "uuid" {
      b.UserType = &Uuid{}
      n += copy(b.UserType[:], buf[n:])
   }
   return n, nil
}

