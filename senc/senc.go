package senc

import (
   "154.pages.dev/sofia"
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "io"
)

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
   Sample       []Sample
}

type Sample struct {
   InitializationVector uint64
   SubsampleCount       uint16
   Subsample           []Subsample
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

///

func (e *Sample) read(src io.Reader, box *Box) error {
   err := binary.Read(src, binary.BigEndian, &e.InitializationVector)
   if err != nil {
      return err
   }
   if box.senc_use_subsamples() {
      err := binary.Read(src, binary.BigEndian, &e.SubsampleCount)
      if err != nil {
         return err
      }
      e.Subsample = make([]Subsample, e.SubsampleCount)
      for i, sub := range e.Subsample {
         err := sub.read(src)
         if err != nil {
            return err
         }
         e.Subsample[i] = sub
      }
   }
   return nil
}

func (e Sample) write(w io.Writer, box Box) error {
   err := binary.Write(w, binary.BigEndian, e.InitializationVector)
   if err != nil {
      return err
   }
   if box.senc_use_subsamples() {
      err := binary.Write(w, binary.BigEndian, e.SubsampleCount)
      if err != nil {
         return err
      }
      for _, sub := range e.Subsample {
         err := sub.write(w)
         if err != nil {
            return err
         }
      }
   }
   return nil
}

func (s *Box) read(src io.Reader) error {
   err := s.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &s.SampleCount)
   if err != nil {
      return err
   }
   s.Sample = make([]Sample, s.SampleCount)
   for i, value := range s.Sample {
      err := value.read(src, s)
      if err != nil {
         return err
      }
      s.Sample[i] = value
   }
   return nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (s Box) senc_use_subsamples() bool {
   return s.FullBoxHeader.GetFlags()&2 >= 1
}

func (s Box) write(w io.Writer) error {
   err := s.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   err = s.FullBoxHeader.Write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.SampleCount)
   if err != nil {
      return err
   }
   for _, value := range s.Sample {
      err := value.write(w, s)
      if err != nil {
         return err
      }
   }
   return nil
}

func (s *Subsample) read(src io.Reader) error {
   return binary.Read(src, binary.BigEndian, s)
}

func (s Subsample) write(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, s)
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (e Sample) DecryptCenc(text, key []byte) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   var iv [16]byte
   binary.BigEndian.PutUint64(iv[:], e.InitializationVector)
   stream := cipher.NewCTR(block, iv[:])
   if len(e.Subsample) >= 1 {
      var pos uint32
      for _, ss := range e.Subsample {
         nrClear := uint32(ss.BytesOfClearData)
         if nrClear >= 1 {
            pos += nrClear
         }
         nrEnc := ss.BytesOfProtectedData
         if nrEnc >= 1 {
            stream.XORKeyStream(text[pos:pos+nrEnc], text[pos:pos+nrEnc])
            pos += nrEnc
         }
      }
   } else {
      stream.XORKeyStream(text, text)
   }
   return nil
}
