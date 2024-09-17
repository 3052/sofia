package saio

import (
   "154.pages.dev/sofia"
   "encoding/binary"
)

func (b *Box) Read(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   if b.FullBoxHeader.GetFlags() & 1 >= 1 {
      n, err = binary.Decode(buf, binary.BigEndian, &b.AuxInfoType)
      if err != nil {
         return err
      }
      buf = buf[n:]
      n, err = binary.Decode(buf, binary.BigEndian, &b.AuxInfoTypeParameter)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   n, err = binary.Decode(buf, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   buf = buf[n:]
   if b.FullBoxHeader.Version == 0 {
      n = 4
   } else {
      n = 8
   }
   b.Offset = make([][]byte, b.EntryCount)
   for i := range b.Offset {
      b.Offset[i] = buf[:n]
      buf = buf[n:]
   }
   return nil
}

// ISO/IEC 14496-12
//
//   aligned(8) class SampleAuxiliaryInformationOffsetsBox extends FullBox(
//      'saio', version, flags
//   ) {
//      if (flags & 1) {
//         unsigned int(32) aux_info_type;
//         unsigned int(32) aux_info_type_parameter;
//      }
//      unsigned int(32) entry_count;
//      if ( version == 0 ) {
//         unsigned int(32) offset[ entry_count ];
//      } else {
//         unsigned int(64) offset[ entry_count ];
//      }
//   }
type Box struct {
   BoxHeader            sofia.BoxHeader
   FullBoxHeader        sofia.FullBoxHeader
   AuxInfoType          uint32
   AuxInfoTypeParameter uint32
   EntryCount           uint32
   Offset               [][]byte
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   if b.FullBoxHeader.GetFlags()&1 >= 1 {
      buf = binary.BigEndian.AppendUint32(buf, b.AuxInfoType)
      buf = binary.BigEndian.AppendUint32(buf, b.AuxInfoTypeParameter)
   }
   buf = binary.BigEndian.AppendUint32(buf, b.EntryCount)
   for _, offset := range b.Offset {
      buf = append(buf, offset...)
   }
   return buf, nil
}
