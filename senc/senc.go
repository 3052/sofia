package senc

import (
   "154.pages.dev/sofia"
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "io"
)

func (b *Box) Read(src io.Reader) error {
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   b.Sample = make([]Sample, b.SampleCount)
   for i, value := range b.Sample {
      value.box = b
      err := value.read(src)
      if err != nil {
         return err
      }
      b.Sample[i] = value
   }
   return nil
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

// senc_use_subsamples: flag mask is 0x000002.
func (b Box) senc_use_subsamples() bool {
   return b.FullBoxHeader.GetFlags()&2 >= 1
}

func (s *Subsample) read(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, s)
}

func (s Subsample) write(dst io.Writer) error {
   return binary.Write(dst, binary.BigEndian, s)
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (s Sample) DecryptCenc(text, key []byte) error {
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

func (b Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.SampleCount)
   if err != nil {
      return err
   }
   for _, value := range b.Sample {
      err := value.write(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

func (s *Sample) read(src io.Reader) error {
   err := binary.Read(src, binary.BigEndian, &s.InitializationVector)
   if err != nil {
      return err
   }
   if s.box.senc_use_subsamples() {
      err := binary.Read(src, binary.BigEndian, &s.SubsampleCount)
      if err != nil {
         return err
      }
      s.Subsample = make([]Subsample, s.SubsampleCount)
      for i, sub := range s.Subsample {
         err := sub.read(src)
         if err != nil {
            return err
         }
         s.Subsample[i] = sub
      }
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

func (s Sample) write(dst io.Writer) error {
   err := binary.Write(dst, binary.BigEndian, s.InitializationVector)
   if err != nil {
      return err
   }
   if s.box.senc_use_subsamples() {
      err := binary.Write(dst, binary.BigEndian, s.SubsampleCount)
      if err != nil {
         return err
      }
      for _, sub := range s.Subsample {
         err := sub.write(dst)
         if err != nil {
            return err
         }
      }
   }
   return nil
}
