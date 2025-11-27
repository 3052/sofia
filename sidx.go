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

// AddReference appends a new reference entry to the SidxBox.
func (b *SidxBox) AddReference(referencedSize, duration uint32) error {
   if referencedSize > 0x7FFFFFFF {
      return errors.New("referencedSize exceeds maximum 31-bit value")
   }
   // The reference_count field in the box is uint16 (max 65535).
   if len(b.References) >= 65535 {
      return errors.New("maximum number of sidx references (65535) exceeded")
   }
   ref := SidxReference{
      ReferencedSize:     referencedSize,
      SubsegmentDuration: duration,
      // Defaulting StartsWithSAP to true and SAPType to 1 is crucial.
      // Otherwise, players may treat the segment as non-seekable.
      StartsWithSAP: true,
      SAPType:       1,
   }
   b.References = append(b.References, ref)
   return nil
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

   // Skip reserved (2 bytes)
   offset += 2

   referenceCount := binary.BigEndian.Uint16(data[offset : offset+2])
   offset += 2

   b.References = make([]SidxReference, referenceCount)

   for i := 0; i < int(referenceCount); i++ {
      if len(data) < offset+12 {
         return fmt.Errorf("sidx box is truncated at reference index %d", i)
      }

      // Word 1: ReferenceType (1 bit) | ReferencedSize (31 bits)
      val1 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].ReferenceType = (val1 >> 31) == 1
      b.References[i].ReferencedSize = val1 & 0x7FFFFFFF
      offset += 4

      // Word 2: SubsegmentDuration
      b.References[i].SubsegmentDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4

      // Word 3: StartsWithSAP (1 bit) | SAPType (3 bits) | SAPDeltaTime (28 bits)
      val2 := binary.BigEndian.Uint32(data[offset : offset+4])
      b.References[i].StartsWithSAP = (val2 >> 31) == 1
      b.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      b.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
      offset += 4
   }

   return nil
}

// Encode serializes the SidxBox to a byte slice.
func (b *SidxBox) Encode() []byte {
   // Size calculation:
   // Header(8) + Ver/Flags(4) + ID(4) + Time(4) + Res(2) + Count(2) = 24 fixed
   // + EPT/Offset (8 for v0, 16 for v1)
   // + References (12 each)
   size := 24 + (len(b.References) * 12)
   if b.Version == 0 {
      size += 8
   } else {
      size += 16
   }

   b.Header.Size = uint32(size)
   buf := make([]byte, size)

   headerBytes := b.Header.Encode()
   copy(buf[0:8], headerBytes)

   buf[8] = b.Version
   binary.BigEndian.PutUint32(buf[8:12], b.Flags&0x00FFFFFF|uint32(b.Version)<<24)

   offset := 12
   binary.BigEndian.PutUint32(buf[offset:offset+4], b.ReferenceID)
   offset += 4
   binary.BigEndian.PutUint32(buf[offset:offset+4], b.Timescale)
   offset += 4

   if b.Version == 0 {
      binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(b.EarliestPresentationTime))
      offset += 4
      binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(b.FirstOffset))
      offset += 4
   } else {
      binary.BigEndian.PutUint64(buf[offset:offset+8], b.EarliestPresentationTime)
      offset += 8
      binary.BigEndian.PutUint64(buf[offset:offset+8], b.FirstOffset)
      offset += 8
   }

   // Reserved (2 bytes) zeroed by alloc
   offset += 2

   binary.BigEndian.PutUint16(buf[offset:offset+2], uint16(len(b.References)))
   offset += 2

   for _, ref := range b.References {
      // Word 1: [RefType 1bit][ReferencedSize 31bit]
      val1 := ref.ReferencedSize & 0x7FFFFFFF
      if ref.ReferenceType {
         val1 |= 1 << 31
      }
      binary.BigEndian.PutUint32(buf[offset:offset+4], val1)
      offset += 4

      // Word 2: Duration
      binary.BigEndian.PutUint32(buf[offset:offset+4], ref.SubsegmentDuration)
      offset += 4

      // Word 3: [StartsWithSAP 1bit][SAPType 3bit][SAPDeltaTime 28bit]
      val2 := ref.SAPDeltaTime & 0x0FFFFFFF
      val2 |= uint32(ref.SAPType&0x07) << 28
      if ref.StartsWithSAP {
         val2 |= 1 << 31
      }
      binary.BigEndian.PutUint32(buf[offset:offset+4], val2)
      offset += 4
   }

   return buf
}
