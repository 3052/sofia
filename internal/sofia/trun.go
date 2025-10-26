package mp4

import (
   "encoding/binary"
   "errors"
)

// SampleInfo holds details about a single sample in a track run.
type SampleInfo struct {
   Size uint32
   // Duration, CompositionTimeOffset, etc., could be added here if needed.
}

// TrunBox represents the 'trun' box (Track Run Box).
type TrunBox struct {
   Header  BoxHeader
   Version byte
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
   trun.Version = data[8]
   trun.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF

   sampleCount := binary.BigEndian.Uint32(data[12:16])
   offset := 16

   // Skip data_offset if present
   if trun.Flags&0x000001 != 0 {
      offset += 4
   }
   // Skip first_sample_flags if present
   if trun.Flags&0x000004 != 0 {
      offset += 4
   }

   trun.Samples = make([]SampleInfo, sampleCount)
   for i := uint32(0); i < sampleCount; i++ {
      // This simplified parser assumes only sample_size is present.
      // A full implementation would check flags for duration, CTO, etc.
      if trun.Flags&0x000200 == 0 {
         return TrunBox{}, errors.New("trun parsing error: expected sample_size_present flag")
      }
      trun.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   return trun, nil
}

func (b *TrunBox) Encode() []byte { return nil } // Omitted for brevity
