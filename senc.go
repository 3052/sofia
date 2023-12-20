package sofia

import (
   "encoding/binary"
   "io"
)

type Subsample struct {
   BytesOfClearData uint16
   BytesOfProtectedData uint32
}

func (s *Subsample) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, s)
}

type Sample struct {
   InitializationVector [8]byte
   Subsample_Count uint16
   Subsamples []Subsample
}

func (s *Sample) Decode(r io.Reader) error {
   _, err := r.Read(s.InitializationVector[:])
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.Subsample_Count)
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
   return nil
}

// aligned(8) class SampleEncryptionBox extends FullBox(
//    'senc',
//    version,
//    flags
// ) {
//    unsigned int(32) sample_count;
//    {
//       unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//       unsigned int(16) subsample_count;
//       {
//          unsigned int(16) BytesOfClearData;
//          unsigned int(32) BytesOfProtectedData;
//       } [subsample_count ]
//    }[ sample_count ]
// }
type SampleEncryptionBox struct {
   Header FullBoxHeader
   Sample_Count uint32
   Samples []Sample
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
      var sam Sample
      err := sam.Decode(r)
      if err != nil {
         return err
      }
      s.Samples = append(s.Samples, sam)
   }
   return nil
}
