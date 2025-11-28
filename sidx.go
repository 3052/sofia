package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

// SidxReference holds the data for a single entry in the sidx list.
type SidxReference struct {
   ReferenceType      bool   // 1 bit: 0 = media, 1 = index
   ReferencedSize     uint32 // 31 bits
   SubsegmentDuration uint32
   StartsWithSAP      bool   // 1 bit
   SAPType            uint8  // 3 bits
   SAPDeltaTime       uint32 // 28 bits
}

// SidxBox represents the 'sidx' box (Segment Index Box).
type SidxBox struct {
   Header                   BoxHeader
   Version                  byte
   Flags                    uint32
   ReferenceID              uint32
   Timescale                uint32
   EarliestPresentationTime uint64
   FirstOffset              uint64
   References               []SidxReference
}

// Parse parses the 'sidx' box from a byte slice.
func (b *SidxBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

   if len(data) < 12 {
      return errors.New("sidx box too short for version and flags")
   }

   b.Version = data[8]
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF

   offset := 12
   if len(data) < offset+8 {
      return errors.New("sidx box is too short for referenceID and timescale")
   }

   b.ReferenceID = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4

   if b.Version == 0 {
      if len(data) < offset+8 {
         return errors.New("sidx v0 box is too short for EPT and first_offset")
      }
      b.EarliestPresentationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.FirstOffset = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   } else {
      if len(data) < offset+16 {
         return errors.New("sidx v1 box is too short for EPT and first_offset")
      }
      b.EarliestPresentationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.FirstOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }

   if len(data) < offset+4 {
      return errors.New("sidx box is too short for reference_count")
   }

   offset += 2 // Skip reserved
   referenceCount := binary.BigEndian.Uint16(data[offset : offset+2])
   offset += 2

   b.References = make([]SidxReference, referenceCount)

   for i := 0; i < int(referenceCount); i++ {
      if len(data) < offset+12 {
         return fmt.Errorf("sidx box is truncated at reference index %d", i)
      }
      val1 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].ReferenceType = (val1 >> 31) == 1
      b.References[i].ReferencedSize = val1 & 0x7FFFFFFF
      offset += 4

      b.References[i].SubsegmentDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4

      val2 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].StartsWithSAP = (val2 >> 31) == 1
      b.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      b.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
      offset += 4
   }
   return nil
}
