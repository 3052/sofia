// encryption.go
package sofia

import (
   "crypto/cipher"
   "errors"
)

// --- PSSH ---
type PsshBox struct {
   Header   *BoxHeader
   Version  byte
   Flags    [3]byte
   SystemID [16]byte
   KIDs     [][16]byte
   Data     []byte
}

func DecodePsshBox(data []byte) (*PsshBox, error) {
   b := &PsshBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 28 { // 8 byte header + 4 byte version/flags + 16 byte systemID
      return nil, errors.New("pssh too short")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])
   copy(b.SystemID[:], p.Bytes(16))

   if b.Version > 0 {
      if len(data) < p.offset+4 {
         return nil, errors.New("pssh too short for KID count")
      }
      kidCount := p.Uint32()
      if len(data) < p.offset+int(kidCount*16) {
         return nil, errors.New("pssh too short for KIDs")
      }
      b.KIDs = make([][16]byte, kidCount)
      for i := 0; i < int(kidCount); i++ {
         copy(b.KIDs[i][:], p.Bytes(16))
      }
   }

   if len(data) < p.offset+4 {
      return nil, errors.New("pssh too short for data size")
   }
   dataSize := p.Uint32()
   if len(data) < p.offset+int(dataSize) {
      return nil, errors.New("pssh size mismatch")
   }
   b.Data = p.Bytes(int(dataSize))
   return b, nil
}

// --- TENC ---
// TencBox defines the Track Encryption Box ('tenc'), which contains
// default encryption parameters for a track.
// Specification: ISO/IEC 23001-7
type TencBox struct {
   Header                 *BoxHeader
   Version                byte
   Flags                  uint32
   DefaultIsProtected     byte
   DefaultPerSampleIVSize byte
   DefaultKID             [16]byte
   DefaultConstantIVSize  byte   // Present if DefaultIsProtected=1 and DefaultPerSampleIVSize=0
   DefaultConstantIV      []byte // Present if DefaultIsProtected=1 and DefaultPerSampleIVSize=0
}

func DecodeTencBox(data []byte) (*TencBox, error) {
   b := &TencBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   p := parser{data: data, offset: 8}
   if len(data) < p.offset+4 {
      return nil, errors.New("tenc box too short for version/flags")
   }
   versionAndFlags := p.Uint32()
   b.Version = byte(versionAndFlags >> 24)
   b.Flags = versionAndFlags & 0x00FFFFFF

   if b.Version == 0 {
      // Based on the error, the reserved field for this file is 2 bytes.
      // Payload: reserved(2) + isProtected(1) + perSampleIVSize(1) + KID(16) = 20 bytes.
      const requiredV0PayloadSize = 20
      if len(data) < p.offset+requiredV0PayloadSize {
         return nil, errors.New("tenc v0 box too short for required fields")
      }

      // Correctly skip the 2 reserved bytes.
      _ = p.Bytes(2)

      b.DefaultIsProtected = p.Byte()
      b.DefaultPerSampleIVSize = p.Byte()
      copy(b.DefaultKID[:], p.Bytes(16))

      if b.DefaultIsProtected == 1 && b.DefaultPerSampleIVSize == 0 {
         if p.offset < int(b.Header.Size) {
            if len(data) < p.offset+1 {
               return nil, errors.New("tenc box truncated before constant IV size")
            }
            b.DefaultConstantIVSize = p.Byte()
            if len(data) < p.offset+int(b.DefaultConstantIVSize) {
               return nil, errors.New("tenc box truncated, not enough data for constant IV")
            }
            b.DefaultConstantIV = p.Bytes(int(b.DefaultConstantIVSize))
         }
      }
   }
   // For other versions, we do nothing and leave the fields as their zero-value.
   return b, nil
}

// --- SENC ---
type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

type SencSample struct {
   IV         []byte
   Subsamples []Subsample
}

type SencBox struct {
   Header  *BoxHeader
   Flags   uint32
   Samples []SencSample
}

func DecodeSencBox(data []byte) (*SencBox, error) {
   b := &SencBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 16 { // 8 byte header, 4 byte flags, 4 byte sample count
      return nil, errors.New("senc too short")
   }

   p := parser{data: data, offset: 8}
   b.Flags = p.Uint32() & 0x00FFFFFF
   sampleCount := p.Uint32()

   b.Samples = make([]SencSample, sampleCount)
   const ivSize = 8
   subsamplesPresent := b.Flags&0x000002 != 0
   for i := uint32(0); i < sampleCount; i++ {
      if len(data) < p.offset+ivSize {
         return nil, errors.New("senc truncated while reading IV")
      }
      b.Samples[i].IV = p.Bytes(ivSize)

      if subsamplesPresent {
         if len(data) < p.offset+2 {
            return nil, errors.New("senc truncated while reading subsample count")
         }
         subsampleCount := p.Uint16()
         b.Samples[i].Subsamples = make([]Subsample, subsampleCount)
         for j := uint16(0); j < subsampleCount; j++ {
            if len(data) < p.offset+6 {
               return nil, errors.New("senc truncated while reading subsample")
            }
            clear := p.Uint16()
            prot := p.Uint32()
            b.Samples[i].Subsamples[j] = Subsample{clear, prot}
         }
      }
   }
   return b, nil
}

// --- Logic ---
func Decrypt(data []byte, sample *SencSample, block cipher.Block) {
   if sample == nil || len(sample.IV) == 0 {
      return
   }
   iv := sample.IV
   if len(iv) == 8 {
      paddedIV := make([]byte, 16)
      copy(paddedIV, iv)
      iv = paddedIV
   }
   stream := cipher.NewCTR(block, iv)
   if len(sample.Subsamples) == 0 {
      stream.XORKeyStream(data, data)
   } else {
      offset := 0
      for _, subsample := range sample.Subsamples {
         offset += int(subsample.BytesOfClearData)
         if subsample.BytesOfProtectedData > 0 {
            end := offset + int(subsample.BytesOfProtectedData)
            if end > len(data) {
               end = len(data)
            }
            chunk := data[offset:end]
            stream.XORKeyStream(chunk, chunk)
            offset = end
         }
      }
   }
}
