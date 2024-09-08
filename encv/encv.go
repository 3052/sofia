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
type VisualSampleEntry struct {
   SampleEntry SampleEntry
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

func (v *VisualSampleEntry) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   err := v.SampleEntry.read(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &v.Extends)
   if err != nil {
      return err
   }
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "sinf":
            _, size := head.GetSize()
            v.ProtectionScheme.BoxHeader = head
            err := v.ProtectionScheme.Read(r, size)
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
            err := value.Read(r)
            if err != nil {
               return err
            }
            v.Boxes = append(v.Boxes, &value)
         default:
            return sofia.Error{v.SampleEntry.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (v VisualSampleEntry) write(w io.Writer) error {
   err := v.SampleEntry.write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, v.Extends)
   if err != nil {
      return err
   }
   for _, value := range v.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   return v.ProtectionScheme.Write(w)
}
