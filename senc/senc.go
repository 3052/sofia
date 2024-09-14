package senc

import (
   "154.pages.dev/sofia"
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
)

// senc_use_subsamples: flag mask is 0x000002.
func (b *Box) senc_use_subsamples() bool {
   return b.FullBoxHeader.GetFlags()&2 >= 1
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (s *Sample) DecryptCenc(text, key []byte) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   var iv [16]byte
   binary.BigEndian.PutUint64(iv[:], s.InitializationVector)
   stream := cipher.NewCTR(block, iv[:])
   if len(s.Subsample) >= 1 {
      var i uint32
      for _, sub := range s.Subsample {
         clear := uint32(sub.BytesOfClearData)
         if clear >= 1 {
            i += clear
         }
         protected := sub.BytesOfProtectedData
         if protected >= 1 {
            stream.XORKeyStream(text[i:i+protected], text[i:i+protected])
            i += protected
         }
      }
   } else {
      stream.XORKeyStream(text, text)
   }
   return nil
}

// ISO/IEC 23001-7
//
// if the version of the SampleEncryptionBox is 0 and the flag
// senc_use_subsamples is set, UseSubSampleEncryption is set to 1
//
//   aligned(8) class SampleEncryptionBox extends FullBox(
//      'senc', version, flags
//   ) {
//      unsigned int(32) sample_count;
//      {
//         unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//         if (UseSubSampleEncryption) {
//            unsigned int(16) subsample_count;
//            {
//               unsigned int(16) BytesOfClearData;
//               unsigned int(32) BytesOfProtectedData;
//            } [subsample_count ]
//         }
//      }[ sample_count ]
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   SampleCount   uint32
   Sample        []Sample
}

type Sample struct {
   InitializationVector uint64
   SubsampleCount       uint16
   Subsample            []Subsample
   box                  *Box
}

func (s Subsample) Append(buf []byte) ([]byte, error) {
   return binary.Append(buf, binary.BigEndian, s)
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

func (s *Sample) Append(buf []byte) ([]byte, error) {
   buf, err := binary.Append(buf, binary.BigEndian, s.InitializationVector)
   if err != nil {
      return nil, err
   }
   if s.box.senc_use_subsamples() {
      buf, err = binary.Append(buf, binary.BigEndian, s.SubsampleCount)
      if err != nil {
         return nil, err
      }
      for _, sub := range s.Subsample {
         buf, err = sub.Append(buf)
         if err != nil {
            return nil, err
         }
      }
   }
   return buf, nil
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.SampleCount)
   if err != nil {
      return nil, err
   }
   for _, sam := range b.Sample {
      buf, err = sam.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

func (s *Subsample) Decode(buf []byte) (int, error) {
   return binary.Decode(buf, binary.BigEndian, s)
}

func (s *Sample) Decode(buf []byte) (int, error) {
   ns, err := binary.Decode(buf, binary.BigEndian, &s.InitializationVector)
   if err != nil {
      return 0, err
   }
   if s.box.senc_use_subsamples() {
      n, err := binary.Decode(buf[ns:], binary.BigEndian, &s.SubsampleCount)
      if err != nil {
         return 0, err
      }
      ns += n
      s.Subsample = make([]Subsample, s.SubsampleCount)
      for i, sub := range s.Subsample {
         n, err = sub.Decode(buf[ns:])
         if err != nil {
            return 0, err
         }
         ns += n
         s.Subsample[i] = sub
      }
   }
   return ns, nil
}

func (b *Box) Decode(buf []byte) error {
   ns, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   n, err := binary.Decode(buf[ns:], binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   ns += n
   b.Sample = make([]Sample, b.SampleCount)
   for i, sam := range b.Sample {
      sam.box = b
      n, err = sam.Decode(buf[ns:])
      if err != nil {
         return err
      }
      ns += n
      b.Sample[i] = sam
   }
   return nil
}
