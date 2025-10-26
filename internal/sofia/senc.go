package mp4

import (
   "encoding/binary"
   "encoding/hex"
   "fmt"
   "log"
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

// ParseSenc parses the 'senc' box from a byte slice.
func ParseSenc(data []byte) (SencBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SencBox{}, err
   }
   var senc SencBox
   senc.Header = header
   senc.RawData = data[:header.Size]

   log.Printf("[PARSE SENC] Box size: %d", len(data))
   senc.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   log.Printf("[PARSE SENC] Flags: 0x%06x", senc.Flags)

   offset := 12
   if offset+4 > len(data) {
      return SencBox{}, fmt.Errorf("senc box too short for sample count")
   }
   sampleCount := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   log.Printf("[PARSE SENC] Sample count: %d", sampleCount)

   senc.Samples = make([]SampleEncryptionInfo, sampleCount)
   const ivSize = 8
   subsamplesPresent := senc.Flags&0x000002 != 0

   for i := uint32(0); i < sampleCount; i++ {
      if offset+ivSize > len(data) {
         return SencBox{}, fmt.Errorf("senc box truncated at IV for sample %d", i)
      }
      senc.Samples[i].IV = data[offset : offset+ivSize]
      offset += ivSize
      log.Printf("[PARSE SENC] Sample %d: IV=%s", i, hex.EncodeToString(senc.Samples[i].IV))

      if subsamplesPresent {
         if offset+2 > len(data) {
            return SencBox{}, fmt.Errorf("senc box truncated at subsample count for sample %d", i)
         }
         subsampleCount := binary.BigEndian.Uint16(data[offset : offset+2])
         offset += 2
         senc.Samples[i].Subsamples = make([]SubsampleInfo, subsampleCount)
         log.Printf("[PARSE SENC] Sample %d: Subsample count=%d", i, subsampleCount)

         for j := uint16(0); j < subsampleCount; j++ {
            if offset+6 > len(data) {
               return SencBox{}, fmt.Errorf("senc box truncated at subsample data for sample %d", i)
            }
            clearBytes := binary.BigEndian.Uint16(data[offset : offset+2])
            protectedBytes := binary.BigEndian.Uint32(data[offset+2 : offset+6])
            senc.Samples[i].Subsamples[j] = SubsampleInfo{
               BytesOfClearData:     clearBytes,
               BytesOfProtectedData: protectedBytes,
            }
            offset += 6
            log.Printf("[PARSE SENC] Sample %d, Subsample %d: Clear=%-5d Protected=%-5d", i, j, clearBytes, protectedBytes)
         }
      }
   }
   return senc, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *SencBox) Encode() []byte {
   return b.RawData
}
