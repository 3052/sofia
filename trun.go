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
   Header           BoxHeader
   Flags            uint32
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Samples          []SampleInfo
}

func (b *TrunBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   b.SampleCount = binary.BigEndian.Uint32(data[12:16])
   offset := 16

   // Data Offset Present
   if b.Flags&0x000001 != 0 {
      if offset+4 > len(data) {
         return errors.New("trun too short for data offset")
      }
      b.DataOffset = int32(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }

   // First Sample Flags Present (0x04)
   if b.Flags&0x000004 != 0 {
      if offset+4 > len(data) {
         return errors.New("trun too short for first sample flags")
      }
      b.FirstSampleFlags = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }

   // Calculate size of one sample entry
   sampleEntrySize := 0
   if b.Flags&0x000100 != 0 {
      sampleEntrySize += 4
   } // Duration
   if b.Flags&0x000200 != 0 {
      sampleEntrySize += 4
   } // Size
   if b.Flags&0x000400 != 0 {
      sampleEntrySize += 4
   } // Flags
   if b.Flags&0x000800 != 0 {
      sampleEntrySize += 4
   } // CTO

   // Safety check
   if len(data)-offset < int(b.SampleCount)*sampleEntrySize {
      return errors.New("trun box too short for declared samples")
   }

   b.Samples = make([]SampleInfo, b.SampleCount)
   for i := uint32(0); i < b.SampleCount; i++ {
      if b.Flags&0x000100 != 0 {
         b.Samples[i].Duration = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000200 != 0 {
         b.Samples[i].Size = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000400 != 0 {
         b.Samples[i].Flags = binary.BigEndian.Uint32(data[offset : offset+4])
         offset += 4
      }
      if b.Flags&0x000800 != 0 {
         val := binary.BigEndian.Uint32(data[offset : offset+4])
         b.Samples[i].CompositionTimeOffset = int32(val)
         offset += 4
      }
   }
   return nil
}
