package sofia

import (
   "encoding/binary"
   "encoding/hex"
   "fmt"
   "io"
   "log"
)

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

func (f *FullBoxHeader) GetFlags() uint32 {
   var flag [4]byte
   copy(flag[1:], f.Flags[:])
   return binary.BigEndian.Uint32(flag[:])
}

func (u *Uuid) String() string {
   return hex.EncodeToString(u[:])
}

type Uuid [16]uint8

func (s *SampleEntry) Decode(data []byte) (int, error) {
   n := 6 // reserved
   n1, err := binary.Decode(data[n:], binary.BigEndian, &s.DataReferenceIndex)
   if err != nil {
      return 0, err
   }
   n += n1
   return n, nil
}

// ISO/IEC 14496-12
//
//   aligned(8) abstract class SampleEntry(
//      unsigned int(32) format
//   ) extends Box(format) {
//      const unsigned int(8)[6] reserved = 0;
//      unsigned int(16) data_reference_index;
//   }
type SampleEntry struct {
   BoxHeader          BoxHeader
   _                  [6]uint8
   DataReferenceIndex uint16
}

const PiffExtendedType = "a2394f525a9b4f14a2446c427c648df4"

var Debug = log.New(io.Discard, "", 0)

// ISO/IEC 14496-12
//
//   aligned(8) class Box(
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

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   return append(data, b.Payload...), nil
}

func (b *Box) Read(data []byte) error {
   n, err := b.BoxHeader.Decode(data)
   if err != nil {
      return err
   }
   b.Payload = data[n:b.BoxHeader.Size]
   return nil
}

func (b *BoxError) Error() string {
   return fmt.Sprintf("container:%q box:%q", b.Container.Type, b.Box.Type)
}

type BoxError struct {
   Container BoxHeader
   Box       BoxHeader
}

func (b *BoxHeader) String() string {
   data := fmt.Appendf(nil, "type:%q", b.Type)
   data = fmt.Append(data, " size:", b.Size)
   if b.UserType != nil {
      data = fmt.Append(data, " usertype:", b.UserType)
   }
   return string(data)
}

// ISO/IEC 14496-12
//
//   aligned(8) class BoxHeader(
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

func (s *SampleEntry) Append(data []byte) ([]byte, error) {
   data, err := s.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data = append(data, 0, 0, 0, 0, 0, 0) // reserved
   return binary.BigEndian.AppendUint16(data, s.DataReferenceIndex), nil
}

type Type [4]uint8

func (b *BoxHeader) Append(data []byte) ([]byte, error) {
   data = binary.BigEndian.AppendUint32(data, b.Size)
   data = append(data, b.Type[:]...)
   if b.UserType != nil {
      data = append(data, b.UserType[:]...)
   }
   return data, nil
}

func (t Type) String() string {
   return string(t[:])
}

func (b *BoxHeader) Decode(data []byte) (int, error) {
   n, err := binary.Decode(data, binary.BigEndian, &b.Size)
   if err != nil {
      return 0, err
   }
   n += copy(b.Type[:], data[n:])
   if b.Type.String() == "uuid" {
      b.UserType = &Uuid{}
      n += copy(b.UserType[:], data[n:])
   }
   return n, nil
}
