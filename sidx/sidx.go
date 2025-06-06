package sidx

import (
   "41.neocities.org/sofia"
   "encoding/binary"
)

func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.ReferenceId)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.Timescale)
   if err != nil {
      return err
   }
   data = data[n:]
   if b.FullBoxHeader.Version == 0 {
      n = 4
   } else {
      n = 8
   }
   b.EarliestPresentationTime = data[:n]
   data = data[n:]
   b.FirstOffset = data[:n]
   data = data[n:]
   data = data[2:] // reserved
   n, err = binary.Decode(data, binary.BigEndian, &b.ReferenceCount)
   if err != nil {
      return err
   }
   data = data[n:]
   b.Reference = make([]Reference, 0, b.ReferenceCount)
   for _, refer := range b.Reference {
      n, err = refer.Decode(data)
      if err != nil {
         return err
      }
      data = data[n:]
      b.Reference = append(b.Reference, refer)
   }
   return nil
}

func (r *Reference) SetSize(size uint32) {
   r[0] &= ^r.mask()
   r[0] |= size
}

type Reference [3]uint32

func (r *Reference) Append(data []byte) ([]byte, error) {
   return binary.Append(data, binary.BigEndian, r)
}

func (r *Reference) Decode(data []byte) (int, error) {
   return binary.Decode(data, binary.BigEndian, r)
}

func (*Reference) mask() uint32 {
   return 0xFFFFFFFF >> 1
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r *Reference) Size() uint32 {
   return r[0] & r.mask()
}

// ISO/IEC 14496-12
//
//   aligned(8) class SegmentIndexBox extends FullBox('sidx', version, 0) {
//      unsigned int(32) reference_ID;
//      unsigned int(32) timescale;
//      if (version==0) {
//         unsigned int(32) earliest_presentation_time;
//         unsigned int(32) first_offset;
//      } else {
//         unsigned int(64) earliest_presentation_time;
//         unsigned int(64) first_offset;
//      }
//      unsigned int(16) reserved = 0;
//      unsigned int(16) reference_count;
//      for(i=1; i <= reference_count; i++) {
//         bit (1) reference_type;
//         unsigned int(31) referenced_size;
//         unsigned int(32) subsegment_duration;
//         bit(1) starts_with_SAP;
//         unsigned int(3) SAP_type;
//         unsigned int(28) SAP_delta_time;
//      }
//   }
type Box struct {
   BoxHeader                sofia.BoxHeader
   FullBoxHeader            sofia.FullBoxHeader
   ReferenceId              uint32
   Timescale                uint32
   EarliestPresentationTime []byte
   FirstOffset              []byte
   _                        uint16
   ReferenceCount           uint16
   Reference                []Reference
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, b.FullBoxHeader)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.ReferenceId)
   data = binary.BigEndian.AppendUint32(data, b.Timescale)
   data = append(data, b.EarliestPresentationTime...)
   data = append(data, b.FirstOffset...)
   data = append(data, 0, 0) // reserved
   data = binary.BigEndian.AppendUint16(data, b.ReferenceCount)
   for _, refer := range b.Reference {
      data, err = refer.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return data, nil
}
