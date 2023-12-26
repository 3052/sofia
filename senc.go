package sofia

import (
   "encoding/binary"
   "io"
)

type EncryptionSample struct {
   InitializationVector uint64
   Subsample_Count      uint16
   Subsamples           []Subsample
}

func (e *EncryptionSample) Decode(b *SampleEncryptionBox, r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &e.InitializationVector)
   if err != nil {
      return err
   }
   if b.Senc_Use_Subsamples() {
      err := binary.Read(r, binary.BigEndian, &e.Subsample_Count)
      if err != nil {
         return err
      }
      e.Subsamples = make([]Subsample, e.Subsample_Count)
      for i, sample := range e.Subsamples {
         err := sample.Decode(r)
         if err != nil {
            return err
         }
         e.Subsamples[i] = sample
      }
   }
   return nil
}

func (e EncryptionSample) Encode(b SampleEncryptionBox, w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, e.InitializationVector)
   if err != nil {
      return err
   }
   if b.Senc_Use_Subsamples() {
      err := binary.Write(w, binary.BigEndian, e.Subsample_Count)
      if err != nil {
         return err
      }
      for _, sample := range e.Subsamples {
         err := sample.Encode(w)
         if err != nil {
            return err
         }
      }
   }
   return nil
}

func (b *SampleEncryptionBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &b.Sample_Count)
   if err != nil {
      return err
   }
   b.Samples = make([]EncryptionSample, b.Sample_Count)
   for i, sample := range b.Samples {
      err := sample.Decode(b, r)
      if err != nil {
         return err
      }
      b.Samples[i] = sample
   }
   return nil
}

func (b SampleEncryptionBox) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := b.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, b.Sample_Count); err != nil {
      return err
   }
   for _, sample := range b.Samples {
      err := sample.Encode(b, w)
      if err != nil {
         return err
      }
   }
   return nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (b SampleEncryptionBox) Senc_Use_Subsamples() bool {
   return b.FullBoxHeader.Flags()&2 >= 1
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

func (s *Subsample) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, s)
}

func (s Subsample) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, s)
}

// if the version of the SampleEncryptionBox is 0 and the flag
// senc_use_subsamples is set, UseSubSampleEncryption is set to 1
//
//  aligned(8) class SampleEncryptionBox extends FullBox(
//     'senc', version, flags
//  ) {
//     unsigned int(32) sample_count;
//     {
//        unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//        if (UseSubSampleEncryption) {
//           unsigned int(16) subsample_count;
//           {
//              unsigned int(16) BytesOfClearData;
//              unsigned int(32) BytesOfProtectedData;
//           } [subsample_count ]
//        }
//     }[ sample_count ]
//  }
type SampleEncryptionBox struct {
   BoxHeader     BoxHeader
   FullBoxHeader FullBoxHeader
   Sample_Count  uint32
   Samples       []EncryptionSample
}
