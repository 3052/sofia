package mp4

import (
   "encoding/binary"
   "errors"
   "log"
)

// SampleInfo holds details about a single sample in a track run.
type SampleInfo struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32 // Can be signed
}

// TrunBox represents the 'trun' box (Track Run Box).
type TrunBox struct {
   Header  BoxHeader
   RawData []byte // Stores the original box data for a perfect round trip
   Flags   uint32
   Samples []SampleInfo
}

// ParseTrun parses the 'trun' box from a byte slice.
func ParseTrun(data []byte) (TrunBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrunBox{}, err
   }
   var trun TrunBox
   trun.Header = header
   trun.RawData = data[:header.Size]

   log.Printf("[PARSE TRUN] Box size: %d", len(data))
   trun.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   log.Printf("[PARSE TRUN] Flags: 0x%06x", trun.Flags)

   sampleCount := binary.BigEndian.Uint32(data[12:16])
   log.Printf("[PARSE TRUN] Sample count: %d", sampleCount)
   offset := 16

   if trun.Flags&0x000001 != 0 {
      log.Printf("[PARSE TRUN] Data offset present: %d", binary.BigEndian.Uint32(data[offset:offset+4]))
      offset += 4
   }
   if trun.Flags&0x000004 != 0 {
      log.Printf("[PARSE TRUN] First sample flags present: 0x%x", binary.BigEndian.Uint32(data[offset:offset+4]))
      offset += 4
   }

   trun.Samples = make([]SampleInfo, sampleCount)
   sampleDurationPresent := trun.Flags&0x000100 != 0
   sampleSizePresent := trun.Flags&0x000200 != 0
   sampleFlagsPresent := trun.Flags&0x000400 != 0
   sampleCTOPresent := trun.Flags&0x000800 != 0

   log.Printf("[PARSE TRUN] Per-sample fields present: Duration=%v, Size=%v, Flags=%v, CTO=%v", sampleDurationPresent, sampleSizePresent, sampleFlagsPresent, sampleCTOPresent)

   for i := uint32(0); i < sampleCount; i++ {
      if sampleDurationPresent {
         if offset+4 > len(data) {
            return TrunBox{}, errors.New("trun box truncated at sample duration")
         }
         trun.Samples[i].Duration = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleSizePresent {
         if offset+4 > len(data) {
            return TrunBox{}, errors.New("trun box truncated at sample size")
         }
         trun.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleFlagsPresent {
         if offset+4 > len(data) {
            return TrunBox{}, errors.New("trun box truncated at sample flags")
         }
         trun.Samples[i].Flags = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleCTOPresent {
         if offset+4 > len(data) {
            return TrunBox{}, errors.New("trun box truncated at sample CTO")
         }
         trun.Samples[i].CompositionTimeOffset = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
         offset += 4
      }
      log.Printf("[PARSE TRUN] Sample %d: Duration=%-5d Size=%-5d Flags=0x%08x CTO=%-5d", i, trun.Samples[i].Duration, trun.Samples[i].Size, trun.Samples[i].Flags, trun.Samples[i].CompositionTimeOffset)
   }
   return trun, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TrunBox) Encode() []byte {
   return b.RawData
}
