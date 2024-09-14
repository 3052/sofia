package enca

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

func (s *SampleEntry) Decode(buf []byte, size int64) error {
   buf = buf[:size]
   buf, err = s.SampleEntry.Decode(buf)
   if err != nil {
      return err
   }
   n, err := binary.Decode(buf, binary.BigEndian, &s.Extends)
   if err != nil {
      return err
   }
   buf = buf[n:]
   for len(buf) > 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      buf, err = head.Decode(buf)
      if err != nil {
         return err
      }
      switch head.Type.String() {
      case "sinf":
         n := head.PayloadSize()
         err := s.Sinf.Decode(buf, n)
         if err != nil {
            return err
         }
         buf = buf[n:]
         s.Sinf.BoxHeader = head
      case "btrt", // Criterion
      "dec3", // Hulu
      "esds": // Roku
         box_data := sofia.Box{BoxHeader: head}
         buf, err = box_data.Decode(buf)
         if err != nil {
            return err
         }
         s.Box = append(s.Box, &box_data)
      default:
         return sofia.Error{s.SampleEntry.BoxHeader.Type, head.Type}
      }
   }
   return nil
}

func (s *SampleEntry) Write(dst io.Writer) error {
   err := s.SampleEntry.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, s.Extends)
   if err != nil {
      return err
   }
   for _, value := range s.Box {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   return s.Sinf.Write(dst)
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
