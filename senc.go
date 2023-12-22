package sofia

import (
   "encoding/binary"
   "io"
)

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
      _, err := r.Read(sam.InitializationVector[:])
      if err != nil {
         return err
      }
      if s.Senc_Use_Subsamples() {
         err := binary.Read(r, binary.BigEndian, &sam.Subsample_Count)
         if err != nil {
            return err
         }
         for count := sam.Subsample_Count; count >= 1; count-- {
            var sub Subsample
            err := sub.Decode(r)
            if err != nil {
               return err
            }
            sam.Subsamples = append(sam.Subsamples, sub)
         }
      }
      s.Samples = append(s.Samples, sam)
   }
   return nil
}
func (s SampleEncryptionBox) Senc_Use_Subsamples() bool {
   return s.Header.Flags & 2 >= 1
}

type Subsample struct {
   BytesOfClearData uint16
   BytesOfProtectedData uint32
}

// senc_use_subsamples: flag mask is 0x000002.
// 
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

func (s *Subsample) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, s)
}
