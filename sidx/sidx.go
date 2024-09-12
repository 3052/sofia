package sidx

import (
   "154.pages.dev/sofia"
   "encoding/binary"
   "io"
)

func (b *Box) GetSize() int {
   size := b.BoxHeader.HeaderSize()
   size += binary.Size(b.FullBoxHeader)
   size += binary.Size(b.ReferenceId)
   size += binary.Size(b.Timescale)
   size += binary.Size(b.EarliestPresentationTime)
   size += binary.Size(b.FirstOffset)
   size += binary.Size(b.Reserved)
   size += binary.Size(b.ReferenceCount)
   return size + binary.Size(b.Reference)
}

func (b *Box) Append(size uint32) {
   var ref Reference
   ref.set_referenced_size(size)
   b.Reference = append(b.Reference, ref)
   b.ReferenceCount++
   b.BoxHeader.Size = uint32(b.GetSize())
}

func (b *Box) New() {
   copy(b.BoxHeader.Type[:], "sidx")
}

func (r *Reference) read(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, r)
}

func (b *Box) Read(src io.Reader) error {
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.ReferenceId)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.Timescale)
   if err != nil {
      return err
   }
   if b.FullBoxHeader.Version == 0 {
      b.EarliestPresentationTime = make([]byte, 4)
      b.FirstOffset = make([]byte, 4)
   } else {
      b.EarliestPresentationTime = make([]byte, 8)
      b.FirstOffset = make([]byte, 8)
   }
   _, err = io.ReadFull(src, b.EarliestPresentationTime)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(src, b.FirstOffset)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.Reserved)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.ReferenceCount)
   if err != nil {
      return err
   }
   b.Reference = make([]Reference, b.ReferenceCount)
   for i, value := range b.Reference {
      err := value.read(src)
      if err != nil {
         return err
      }
      b.Reference[i] = value
   }
   return nil
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.ReferenceId)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.Timescale)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.EarliestPresentationTime)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.FirstOffset)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.Reserved)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.ReferenceCount)
   if err != nil {
      return err
   }
   for _, value := range b.Reference {
      err := value.write(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

func (*Reference) mask() uint32 {
   return 0xFFFFFFFF >> 1
}

func (r Reference) write(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, r)
}

// this is the size of the fragment, typically `moof` + `mdat`
func (r Reference) ReferencedSize() uint32 {
   return r[0] & r.mask()
}

func (r Reference) set_referenced_size(size uint32) {
   r[0] &= ^r.mask()
   r[0] |= size
}

type Reference [3]uint32

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
