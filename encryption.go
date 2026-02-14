package sofia

import (
   "crypto/cipher"
   "errors"
)

// --- PSSH ---
type PsshBox struct {
   Header   BoxHeader
   Version  byte
   Flags    [3]byte
   SystemID [16]byte
   KIDs     [][16]byte
   Data     []byte
}

func (b *PsshBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 28 { // 8 byte header + 4 byte version/flags + 16 byte systemID
      return errors.New("pssh too short")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])
   copy(b.SystemID[:], p.Bytes(16))

   if b.Version > 0 {
      if len(data) < p.offset+4 {
         return errors.New("pssh too short for KID count")
      }
      kidCount := p.Uint32()
      if len(data) < p.offset+int(kidCount*16) {
         return errors.New("pssh too short for KIDs")
      }
      b.KIDs = make([][16]byte, kidCount)
      for i := 0; i < int(kidCount); i++ {
         copy(b.KIDs[i][:], p.Bytes(16))
      }
   }

   if len(data) < p.offset+4 {
      return errors.New("pssh too short for data size")
   }
   dataSize := p.Uint32()
   if len(data) < p.offset+int(dataSize) {
      return errors.New("pssh size mismatch")
   }
   b.Data = p.Bytes(int(dataSize))
   return nil
}

// --- SENC ---
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
   if len(data) < 16 { // 8 byte header, 4 byte flags, 4 byte sample count
      return errors.New("senc too short")
   }

   p := parser{data: data, offset: 8}
   b.Flags = p.Uint32() & 0x00FFFFFF
   sampleCount := p.Uint32()

   b.Samples = make([]SampleEncryptionInfo, sampleCount)
   const ivSize = 8
   subsamplesPresent := b.Flags&0x000002 != 0
   for i := uint32(0); i < sampleCount; i++ {
      if len(data) < p.offset+ivSize {
         return errors.New("senc truncated while reading IV")
      }
      b.Samples[i].IV = p.Bytes(ivSize)

      if subsamplesPresent {
         if len(data) < p.offset+2 {
            return errors.New("senc truncated while reading subsample count")
         }
         subsampleCount := p.Uint16()
         b.Samples[i].Subsamples = make([]SubsampleInfo, subsampleCount)
         for j := uint16(0); j < subsampleCount; j++ {
            if len(data) < p.offset+6 {
               return errors.New("senc truncated while reading subsample")
            }
            clear := p.Uint16()
            prot := p.Uint32()
            b.Samples[i].Subsamples[j] = SubsampleInfo{clear, prot}
         }
      }
   }
   return nil
}

// --- Logic ---
func DecryptSample(sample []byte, info *SampleEncryptionInfo, block cipher.Block) {
   if info == nil || len(info.IV) == 0 {
      return
   }
   iv := info.IV
   if len(iv) == 8 {
      paddedIV := make([]byte, 16)
      copy(paddedIV, iv)
      iv = paddedIV
   }
   stream := cipher.NewCTR(block, iv)
   if len(info.Subsamples) == 0 {
      stream.XORKeyStream(sample, sample)
   } else {
      sampleOffset := 0
      for _, subsample := range info.Subsamples {
         sampleOffset += int(subsample.BytesOfClearData)
         if subsample.BytesOfProtectedData > 0 {
            end := sampleOffset + int(subsample.BytesOfProtectedData)
            if end > len(sample) {
               end = len(sample)
            }
            chunk := sample[sampleOffset:end]
            stream.XORKeyStream(chunk, chunk)
            sampleOffset = end
         }
      }
   }
}
