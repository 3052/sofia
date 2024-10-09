package sidx

import (
   "41.neocities.org/sofia"
   "encoding/binary"
)

func (b *Box) GetSize() int {
   size := b.BoxHeader.GetSize()
   size += binary.Size(b.FullBoxHeader)
   size += binary.Size(b.ReferenceId)
   size += binary.Size(b.Timescale)
   size += binary.Size(b.EarliestPresentationTime)
   size += binary.Size(b.FirstOffset)
   size += binary.Size(b.Reserved)
   size += binary.Size(b.ReferenceCount)
   return size + binary.Size(b.Reference)
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
   buf = binary.BigEndian.AppendUint32(buf, b.ReferenceId)
   buf = binary.BigEndian.AppendUint32(buf, b.Timescale)
   buf = append(buf, b.EarliestPresentationTime...)
   buf = append(buf, b.FirstOffset...)
   buf = binary.BigEndian.AppendUint16(buf, b.Reserved)
   buf = binary.BigEndian.AppendUint16(buf, b.ReferenceCount)
   for _, value := range b.Reference {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

// ISO/IEC 14496-12
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
   Reserved                 uint16
   ReferenceCount           uint16
   Reference                []Reference
}

func (b *Box) Read(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.ReferenceId)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.Timescale)
   if err != nil {
      return err
   }
   buf = buf[n:]
   if b.FullBoxHeader.Version == 0 {
      n = 4
   } else {
      n = 8
   }
   b.EarliestPresentationTime = buf[:n]
   buf = buf[n:]
   b.FirstOffset = buf[:n]
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.Reserved)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.ReferenceCount)
   if err != nil {
      return err
   }
   buf = buf[n:]
   b.Reference = make([]Reference, b.ReferenceCount)
   for i, value := range b.Reference {
      n, err = value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[n:]
      b.Reference[i] = value
   }
   return nil
}

func (r *Reference) SetSize(size uint32) {
   (*r)[0] &= ^r.mask()
   (*r)[0] |= size
}

type Reference [3]uint32

func (r Reference) Append(buf []byte) ([]byte, error) {
   return binary.Append(buf, binary.BigEndian, r)
}

func (r *Reference) Decode(buf []byte) (int, error) {
   return binary.Decode(buf, binary.BigEndian, r)
}

func (*Reference) mask() uint32 {
   return 0xFFFFFFFF >> 1
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) Size() uint32 {
   return r[0] & r.mask()
}
