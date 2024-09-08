package encv

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

// ISO/IEC 14496-12
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
   Boxes            []*sofia.Box
   ProtectionScheme sinf.Box
}

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
            s.ProtectionScheme.BoxHeader = head
            err := s.ProtectionScheme.Read(src, size)
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
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            s.Boxes = append(s.Boxes, &value)
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

func (s SampleEntry) Write(dst io.Writer) error {
   err := s.SampleEntry.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, s.Extends)
   if err != nil {
      return err
   }
   for _, value := range s.Boxes {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   return s.ProtectionScheme.Write(dst)
}
