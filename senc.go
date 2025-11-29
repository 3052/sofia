package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

type SubsampleInfo struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

type SampleEncryptionInfo struct {
   IV         []byte
   Subsamples []SubsampleInfo
}

type SencBox struct {
   Header  BoxHeader
   Flags   uint32
   Samples []SampleEncryptionInfo
}

func (b *SencBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   offset := 12
   if offset+4 > len(data) {
      return errors.New("senc too short")
   }
   sampleCount := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   b.Samples = make([]SampleEncryptionInfo, sampleCount)
   const ivSize = 8
   subsamplesPresent := b.Flags&0x000002 != 0
   for i := uint32(0); i < sampleCount; i++ {
      if offset+ivSize > len(data) {
         return fmt.Errorf("senc truncated")
      }
      b.Samples[i].IV = data[offset : offset+ivSize]
      offset += ivSize
      if subsamplesPresent {
         if offset+2 > len(data) {
            return fmt.Errorf("senc truncated")
         }
         cnt := binary.BigEndian.Uint16(data[offset : offset+2])
         offset += 2
         b.Samples[i].Subsamples = make([]SubsampleInfo, cnt)
         for j := uint16(0); j < cnt; j++ {
            if offset+6 > len(data) {
               return fmt.Errorf("senc truncated")
            }
            clear := binary.BigEndian.Uint16(data[offset : offset+2])
            prot := binary.BigEndian.Uint32(data[offset+2 : offset+6])
            b.Samples[i].Subsamples[j] = SubsampleInfo{clear, prot}
            offset += 6
         }
      }
   }
   return nil
}
