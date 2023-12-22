package sofia

import (
   "encoding/binary"
   "io"
)

// senc_use_subsamples: flag mask is 0x000002.
func (s SampleEncryptionBox) Senc_Use_Subsamples() bool {
   return s.Header.Flags & 2 >= 1
}

type Subsample struct {
   BytesOfClearData uint16
   BytesOfProtectedData uint32
}

func (s *Subsample) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, s)
}

// if the version of the SampleEncryptionBox is 0 and the flag
// senc_use_subsamples is set, UseSubSampleEncryption is set to 1
// 
// aligned(8) class SampleEncryptionBox extends FullBox(
//    'senc',
//    version,
//    flags
// ) {
//    unsigned int(32) sample_count;
//    {
//       unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//       if (UseSubSampleEncryption) {
//          unsigned int(16) subsample_count;
//          {
//             unsigned int(16) BytesOfClearData;
//             unsigned int(32) BytesOfProtectedData;
//          } [subsample_count ]
//       }
//    }[ sample_count ]
// }
type SampleEncryptionBox struct {
   Header FullBoxHeader
   Sample_Count uint32
   Samples []SampleEncryption
}

type SampleEncryption struct {
   InitializationVector [8]byte
   Subsample_Count uint16
   Subsamples []Subsample
}

func (s *SampleEncryption) Decode(b *SampleEncryptionBox, r io.Reader) error {
   _, err := r.Read(s.InitializationVector[:])
   if err != nil {
      return err
   }
   if b.Senc_Use_Subsamples() {
      err := binary.Read(r, binary.BigEndian, &s.Subsample_Count)
      if err != nil {
         return err
      }
      for count := s.Subsample_Count; count >= 1; count-- {
         var sub Subsample
         err := sub.Decode(r)
         if err != nil {
            return err
         }
         s.Subsamples = append(s.Subsamples, sub)
      }
   }
   return nil
}

func (s *SampleEncryptionBox) Decode(r io.Reader) error {
   err := s.Header.Decode(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.Sample_Count)
   if err != nil {
      return err
   }
   for count := s.Sample_Count; count >= 1; count-- {
      var sam SampleEncryption
      err := sam.Decode(s, r)
      if err != nil {
         return err
      }
      s.Samples = append(s.Samples, sam)
   }
   return nil
}
