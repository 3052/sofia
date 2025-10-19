package enca

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/sinf"
   "encoding/binary"
)

func (s *SampleEntry) Read(data []byte) error {
   n, err := s.SampleEntry.Decode(data)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &s.S)
   if err != nil {
      return err
   }
   data = data[n:]
   for len(data) > 1 {
      var box sofia.Box
      err := box.Read(data)
      if err != nil {
         return err
      }
      data = data[box.BoxHeader.Size:]
      switch box.BoxHeader.Type.String() {
      case
         // cineMember-avc1
         // criterion-mp4a
         // mubi-avc1
         // rtbf-avc1
         "btrt",
         // hbomax-ec-3
         "dec3",
         // amc-mp4a
         // criterion-mp4a
         // mubi-mp4a
         // nbc-mp4a
         // paramount-mp4a
         // roku-mp4a
         "esds":
         s.Box = append(s.Box, &box)
      case "sinf":
         s.Sinf.BoxHeader = box.BoxHeader
         err = s.Sinf.Read(box.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{s.SampleEntry.BoxHeader, box.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
//
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
   S           struct {
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
   data, err = binary.Append(data, binary.BigEndian, s.S)
   if err != nil {
      return nil, err
   }
   for _, box := range s.Box {
      data, err = box.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return s.Sinf.Append(data)
}
