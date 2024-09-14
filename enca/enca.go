package enca

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
)

func (s *SampleEntry) Append(buf []byte) ([]byte, error) {
   buf, err := s.SampleEntry.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, s.Extends)
   if err != nil {
      return nil, err
   }
   for _, box_data := range s.Box {
      buf, err = box_data.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return s.Sinf.Append(buf)
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

func (s *SampleEntry) Decode(buf []byte) error {
   n, err := s.SampleEntry.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &s.Extends)
   if err != nil {
      return err
   }
   buf = buf[n:]
   for len(buf) > 1 {
      var value sofia.Box
      err := value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "sinf":
         s.Sinf.BoxHeader = value.BoxHeader
         err = s.Sinf.Decode(value.Payload)
         if err != nil {
            return err
         }
      case "btrt", // Criterion
         "dec3", // Hulu
         "esds": // Roku
         s.Box = append(s.Box, &value)
      default:
         return &sofia.Error{s.SampleEntry.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
