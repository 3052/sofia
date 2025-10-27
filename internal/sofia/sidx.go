package mp4

import (
   "encoding/binary"
   "errors"
)

// SidxReference holds the data for a single entry in the sidx list.
type SidxReference struct {
   ReferenceType      bool   // 1 bit
   ReferencedSize     uint32 // 31 bits
   SubsegmentDuration uint32
   StartsWithSAP      bool   // 1 bit
   SAPType            uint8  // 3 bits
   SAPDeltaTime       uint32 // 28 bits
}

// SidxBox represents the 'sidx' box (Segment Index Box).
type SidxBox struct {
   Header                   BoxHeader
   RawData                  []byte
   Version                  byte
   Flags                    uint32
   ReferenceID              uint32
   Timescale                uint32
   EarliestPresentationTime uint64
   FirstOffset              uint64
   References               []SidxReference
}

// ParseSidx parses the 'sidx' box from a byte slice.
func ParseSidx(data []byte) (SidxBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SidxBox{}, err
   }
   var sidx SidxBox
   sidx.Header = header
   sidx.RawData = data[:header.Size]

   sidx.Version = data[8]
   sidx.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   offset := 12

   if len(data) < offset+8 {
      return SidxBox{}, errors.New("sidx box is too short for referenceID and timescale")
   }
   sidx.ReferenceID = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   sidx.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4

   if sidx.Version == 0 {
      if len(data) < offset+8 {
         return SidxBox{}, errors.New("sidx v0 box is too short for EPT and first_offset")
      }
      sidx.EarliestPresentationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      sidx.FirstOffset = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   } else {
      if len(data) < offset+16 {
         return SidxBox{}, errors.New("sidx v1 box is too short for EPT and first_offset")
      }
      sidx.EarliestPresentationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      sidx.FirstOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }

   if len(data) < offset+4 {
      return SidxBox{}, errors.New("sidx box is too short for reference_count")
   }
   // Skip 2 reserved bytes
   offset += 2
   referenceCount := binary.BigEndian.Uint16(data[offset : offset+2])
   offset += 2

   sidx.References = make([]SidxReference, referenceCount)
   for i := 0; i < int(referenceCount); i++ {
      if len(data) < offset+12 {
         return SidxBox{}, errors.New("sidx box is truncated in reference loop")
      }
      val1 := binary.BigEndian.Uint32(data[offset : offset+4])
      sidx.References[i].ReferenceType = (val1 >> 31) == 1
      sidx.References[i].ReferencedSize = val1 & 0x7FFFFFFF
      offset += 4

      sidx.References[i].SubsegmentDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4

      val2 := binary.BigEndian.Uint32(data[offset : offset+4])
      sidx.References[i].StartsWithSAP = (val2 >> 31) == 1
      sidx.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      sidx.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
      offset += 4
   }

   return sidx, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *SidxBox) Encode() []byte {
   return b.RawData
}
