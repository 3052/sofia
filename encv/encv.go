package encv

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
)

// ISO/IEC 14496-12
//
//   class VisualSampleEntry(codingname) extends SampleEntry(codingname) {
//      unsigned int(16) pre_defined = 0;
//      const unsigned int(16) reserved = 0;
//      unsigned int(32)[3] pre_defined = 0;
//      unsigned int(16) width;
//      unsigned int(16) height;
//      template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
//      template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
//      const unsigned int(32) reserved = 0;
//      template unsigned int(16) frame_count = 1;
//      uint(8)[32] compressorname;
//      template unsigned int(16) depth = 0x0018;
//      int(16) pre_defined = -1;
//      // other boxes from derived specifications
//      CleanApertureBox clap; // optional
//      PixelAspectRatioBox pasp; // optional
//   }
type SampleEntry struct {
   SampleEntry sofia.SampleEntry
   Extends     struct {
      _               uint16
      _               uint16
      _               [3]uint32
      Width           uint16
      Height          uint16
      HorizResolution uint32
      VertResolution  uint32
      _               uint32
      FrameCount      uint16
      CompressorName  [32]uint8
      Depth           uint16
      _               int16
   }
   Box  []*sofia.Box
   Sinf sinf.Box
}

func (s *SampleEntry) Append(buf []byte) ([]byte, error) {
   buf, err := s.SampleEntry.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, s.Extends)
   if err != nil {
      return nil, err
   }
   for _, sofia_box := range s.Box {
      buf, err = sofia_box.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return s.Sinf.Append(buf)
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
   for len(buf) >= 1 {
      var sof sofia.Box
      err := sof.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[sof.BoxHeader.Size:]
      switch sof.BoxHeader.Type.String() {
      case "sinf":
         s.Sinf.BoxHeader = sof.BoxHeader
         err := s.Sinf.Decode(sof.Payload)
         if err != nil {
            return err
         }
      case "avcC", // Roku
         "btrt", // Mubi
         "clli", // Max
         "colr", // Paramount
         "dvcC", // Max
         "dvvC", // Max
         "hvcC", // Hulu
         "mdcv", // Max
         "pasp": // Roku
         s.Box = append(s.Box, &sof)
      default:
         return &sofia.Error{s.SampleEntry.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
}
