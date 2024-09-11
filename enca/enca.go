package enca

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

func (s *SampleEntry) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   err := s.SampleEntry.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &s.Extends)
   if err != nil {
      return err
   }
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "sinf":
            _, size := head.GetSize()
            s.Sinf.BoxHeader = head
            err := s.Sinf.Read(src, size)
            if err != nil {
               return err
            }
         case "btrt", // Criterion
            "dec3", // Hulu
            "esds": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            s.Box = append(s.Box, &value)
         default:
            return sofia.Error{s.SampleEntry.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
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
