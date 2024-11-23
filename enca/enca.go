package enca

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/sinf"
   "encoding/binary"
)

// ISO/IEC 14496-12
//   class AudioSampleEntry(codingname) extends SampleEntry(codingname) {
//      const unsigned int(32)[2] reserved = 0;
//      unsigned int(16) channelcount;
//      template unsigned int(16) samplesize = 16;
//      unsigned int(16) pre_defined = 0;
//      const unsigned int(16) reserved = 0 ;
//      template unsigned int(32) samplerate = { default samplerate of media}<<16;
//   }
type SampleEntry struct {
   SampleEntry sofia.SampleEntry
   Extends     struct {
      _            [2]uint32
      ChannelCount uint16
      SampleSize   uint16
      PreDefined   uint16
      _            uint16
      SampleRate   uint32
   }
   Box  []*sofia.Box
   Sinf sinf.Box
}

func (s *SampleEntry) Append(data []byte) ([]byte, error) {
   data, err := s.SampleEntry.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, s.Extends)
   if err != nil {
      return nil, err
   }
   for _, value := range s.Box {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return s.Sinf.Append(data)
}

func (s *SampleEntry) Read(data []byte) error {
   n, err := s.SampleEntry.Decode(data)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &s.Extends)
   if err != nil {
      return err
   }
   data = data[n:]
   for len(data) > 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      data = data[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "btrt", // Criterion
         "dec3", // Hulu
         "esds": // Roku
         s.Box = append(s.Box, &value)
      case "sinf":
         s.Sinf.BoxHeader = value.BoxHeader
         err = s.Sinf.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{s.SampleEntry.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
