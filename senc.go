package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

// SubsampleInfo defines the size of clear and protected data blocks.
type SubsampleInfo struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

// SampleEncryptionInfo contains the IV and subsample data for one sample.
type SampleEncryptionInfo struct {
   IV         []byte
   Subsamples []SubsampleInfo
}

// SencBox represents the 'senc' box (Sample Encryption Box).
type SencBox struct {
   Header  BoxHeader
   RawData []byte
   Flags   uint32
   Samples []SampleEncryptionInfo
}

// Parse parses the 'senc' box from a byte slice.
func (b *SencBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   offset := 12
   if offset+4 > len(data) {
      return errors.New("senc box too short for sample count")
   }
   sampleCount := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   b.Samples = make([]SampleEncryptionInfo, sampleCount)
   const ivSize = 8
   subsamplesPresent := b.Flags&0x000002 != 0
   for i := uint32(0); i < sampleCount; i++ {
      if offset+ivSize > len(data) {
         return fmt.Errorf("senc box truncated at IV for sample %d", i)
      }
      b.Samples[i].IV = data[offset : offset+ivSize]
      offset += ivSize
      if subsamplesPresent {
         if offset+2 > len(data) {
            return fmt.Errorf("senc box truncated at subsample count for sample %d", i)
         }
         subsampleCount := binary.BigEndian.Uint16(data[offset : offset+2])
         offset += 2
         b.Samples[i].Subsamples = make([]SubsampleInfo, subsampleCount)
         for j := uint16(0); j < subsampleCount; j++ {
            if offset+6 > len(data) {
               return fmt.Errorf("senc box truncated at subsample data for sample %d", i)
            }
            clearBytes := binary.BigEndian.Uint16(data[offset : offset+2])
            protectedBytes := binary.BigEndian.Uint32(data[offset+2 : offset+6])
            b.Samples[i].Subsamples[j] = SubsampleInfo{
               BytesOfClearData:     clearBytes,
               BytesOfProtectedData: protectedBytes,
            }
            offset += 6
         }
      }
   }
   return nil
}
