package sofia

import (
   "crypto/cipher"
   "encoding/binary"
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
   if len(data) < 12 {
      return errors.New("pssh too short")
   }
   b.Version = data[8]
   copy(b.Flags[:], data[9:12])
   offset := 12
   if len(data) < offset+16 {
      return errors.New("pssh too short")
   }
   copy(b.SystemID[:], data[offset:offset+16])
   offset += 16
   if b.Version > 0 {
      if len(data) < offset+4 {
         return errors.New("pssh too short")
      }
      kidCount := binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.KIDs = make([][16]byte, kidCount)
      for i := 0; i < int(kidCount); i++ {
         if len(data) < offset+16 {
            return errors.New("pssh too short")
         }
         copy(b.KIDs[i][:], data[offset:offset+16])
         offset += 16
      }
   }
   if len(data) < offset+4 {
      return errors.New("pssh too short")
   }
   dataSize := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   if len(data) < offset+int(dataSize) {
      return errors.New("pssh size mismatch")
   }
   b.Data = data[offset : offset+int(dataSize)]
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
         return errors.New("senc truncated")
      }
      b.Samples[i].IV = data[offset : offset+ivSize]
      offset += ivSize
      if subsamplesPresent {
         if offset+2 > len(data) {
            return errors.New("senc truncated")
         }
         subsampleCount := binary.BigEndian.Uint16(data[offset : offset+2])
         offset += 2
         b.Samples[i].Subsamples = make([]SubsampleInfo, subsampleCount)
         for j := uint16(0); j < subsampleCount; j++ {
            if offset+6 > len(data) {
               return errors.New("senc truncated")
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

// --- Logic ---

// DecryptSample decrypts a single sample in-place.
// info can be nil if the sample is not encrypted.
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
