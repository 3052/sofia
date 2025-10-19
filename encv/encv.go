package encv

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/sinf"
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
   S           struct {
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
   for len(data) >= 1 {
      var box sofia.Box
      err := box.Read(data)
      if err != nil {
         return err
      }
      data = data[box.BoxHeader.Size:]
      switch box.BoxHeader.Type.String() {
      case
         // hboMax-hvc1
         "clli",
         // hboMax-dvh1
         "dvcC",
         // hboMax-hvc1
         "dvvC",
         // hboMax-hvc1
         "mdcv",
         // amc-avc1
         // cineMember-avc1
         // criterion-avc1
         // mubi-avc1
         // nbc-avc1
         // paramount-avc1
         // plex-avc1
         // roku-avc1
         // rtbf-avc1
         // tubi-avc1
         "avcC",
         // cineMember-avc1
         // criterion-mp4a
         // mubi-avc1
         // rtbf-avc1
         "btrt",
         // criterion-avc1
         // paramount-avc1
         "colr",
         // hboMax-dvh1
         // hboMax-hvc1
         "hvcC",
         // nbc-avc1
         // roku-avc1
         "pasp":
         s.Box = append(s.Box, &box)
      case "sinf":
         s.Sinf.BoxHeader = box.BoxHeader
         err := s.Sinf.Read(box.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{s.SampleEntry.BoxHeader, box.BoxHeader}
      }
   }
   return nil
}
