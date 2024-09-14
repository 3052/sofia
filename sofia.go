package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "strconv"
)

func (b *Box) Decode(buf []byte) error {
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

type Error struct {
   Container Type
   Box       Type
}

func (e Error) Error() string {
   buf := []byte("container:")
   buf = strconv.AppendQuote(buf, e.Container.String())
   buf = append(buf, " box type:"...)
   buf = strconv.AppendQuote(buf, e.Box.String())
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

// ISO/IEC 14496-12
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

type Type [4]uint8

func (t Type) String() string {
   return string(t[:])
}

func (u Uuid) String() string {
   return hex.EncodeToString(u[:])
}

type Uuid [16]uint8

func (b *BoxHeader) GetSize() int {
   size := binary.Size(b.Size)
   size += binary.Size(b.Type)
   if b.Type.String() == "uuid" {
      size += binary.Size(b.UserType)
   }
   return size
}

func (b *BoxHeader) Append(buf []byte) ([]byte, error) {
   buf, err := binary.Append(buf, binary.BigEndian, b.Size)
   if err != nil {
      return nil, err
   }
   buf = append(buf, b.Type[:]...)
   if b.Type.String() == "uuid" {
      buf = append(buf, b.UserType[:]...)
   }
   return buf, nil
}

func (b *BoxHeader) Decode(buf []byte) (int, error) {
   n, err := binary.Decode(buf, binary.BigEndian, &b.Size)
   if err != nil {
      return 0, err
   }
   n += copy(b.Type[:], buf[n:])
   if b.Type.String() == "uuid" {
      n += copy(b.UserType[:], buf[n:])
   }
   return n, nil
}

func (s *SampleEntry) Append(buf []byte) ([]byte, error) {
   buf, err := s.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf = append(buf, s.Reserved[:]...)
   return binary.Append(buf, binary.BigEndian, s.DataReferenceIndex)
}

func (f *FullBoxHeader) Decode(buf []byte) (int, error) {
   return binary.Decode(buf, binary.BigEndian, f)
}

func (s *SampleEntry) Decode(buf []byte) (int, error) {
   off := copy(s.Reserved[:], buf)
   n, err := binary.Decode(buf[off:], binary.BigEndian, &s.DataReferenceIndex)
   if err != nil {
      return 0, err
   }
   return off+n, nil
}
