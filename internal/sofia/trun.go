package mp4

import (
   "encoding/binary"
   "errors"
)

// SampleInfo holds details about a single sample in a track run.
type SampleInfo struct {
   Size uint32
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
   trun.RawData = data[:header.Size] // Store the original data

   // Also parse the fields needed for decryption
   trun.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF

   sampleCount := binary.BigEndian.Uint32(data[12:16])
   offset := 16

   if trun.Flags&0x000001 != 0 {
      offset += 4
   } // data_offset_present
   if trun.Flags&0x000004 != 0 {
      offset += 4
   } // first_sample_flags_present

   trun.Samples = make([]SampleInfo, sampleCount)
   sampleDurationPresent := trun.Flags&0x000100 != 0
   sampleSizePresent := trun.Flags&0x000200 != 0
   sampleFlagsPresent := trun.Flags&0x000400 != 0
   sampleCTOPresent := trun.Flags&0x000800 != 0

   for i := uint32(0); i < sampleCount; i++ {
      if sampleDurationPresent {
         offset += 4
      }
      if sampleSizePresent {
         if offset+4 > len(data) {
            return TrunBox{}, errors.New("trun box is truncated while parsing sample sizes")
         }
         trun.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleFlagsPresent {
         offset += 4
      }
      if sampleCTOPresent {
         offset += 4
      }
   }
   return trun, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TrunBox) Encode() []byte {
   return b.RawData
}
