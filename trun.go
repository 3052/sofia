package sofia

import (
   "encoding/binary"
   "errors"
)

type SampleInfo struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32
}

type TrunBox struct {
   Header  BoxHeader
   RawData []byte
   Flags   uint32
   Samples []SampleInfo
}

func (b *TrunBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   sampleCount := binary.BigEndian.Uint32(data[12:16])
   offset := 16
   if b.Flags&0x000001 != 0 {
      offset += 4
   }
   if b.Flags&0x000004 != 0 {
      offset += 4
   }
   b.Samples = make([]SampleInfo, sampleCount)
   sampleDurationPresent := b.Flags&0x000100 != 0
   sampleSizePresent := b.Flags&0x000200 != 0
   sampleFlagsPresent := b.Flags&0x000400 != 0
   sampleCTOPresent := b.Flags&0x000800 != 0
   for i := uint32(0); i < sampleCount; i++ {
      if sampleDurationPresent {
         if offset+4 > len(data) {
            return errors.New("trun box truncated at sample duration")
         }
         b.Samples[i].Duration = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleSizePresent {
         if offset+4 > len(data) {
            return errors.New("trun box truncated at sample size")
         }
         b.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleFlagsPresent {
         if offset+4 > len(data) {
            return errors.New("trun box truncated at sample flags")
         }
         b.Samples[i].Flags = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if sampleCTOPresent {
         if offset+4 > len(data) {
            return errors.New("trun box truncated at sample CTO")
         }
         b.Samples[i].CompositionTimeOffset = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
         offset += 4
      }
   }
   return nil
}
