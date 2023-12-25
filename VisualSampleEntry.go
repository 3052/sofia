package sofia

import (
   "encoding/binary"
   "fmt"
   "io"
)

// aligned(8) abstract class SampleEntry(unsigned int(32) format) extends Box(format) {
//    const unsigned int(8)[6] reserved = 0;
//    unsigned int(16) data_reference_index;
// }
type SampleEntry struct {
   Header  BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
}

func (s *SampleEntry) Decode(r io.Reader) error {
   return binary.Read(r, binary.BigEndian, s)
}

func (s SampleEntry) Encode(w io.Writer) error {
   return binary.Write(w, binary.BigEndian, s)
}

// class VisualSampleEntry(codingname) extends SampleEntry(codingname) {
//    unsigned int(16) pre_defined = 0;
//    const unsigned int(16) reserved = 0;
//    unsigned int(32)[3] pre_defined = 0;
//    unsigned int(16) width;
//    unsigned int(16) height;
//    template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
//    template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
//    const unsigned int(32) reserved = 0;
//    template unsigned int(16) frame_count = 1;
//    uint(8)[32] compressorname;
//    template unsigned int(16) depth = 0x0018;
//    int(16) pre_defined = -1;
//    // other boxes from derived specifications
//    CleanApertureBox clap; // optional
//    PixelAspectRatioBox pasp; // optional
// }
type VisualSampleEntry struct {
   Entry SampleEntry
   Extends struct {
      Pre_Defined uint16
      Reserved uint16
      _ [3]uint32
      Width uint16
      Height uint16
      HorizResolution uint32
      VertResolution uint32
      _ uint32
      Frame_Count uint16
      CompressorName [32]uint8
      Depth uint16
      _ int16
   }
   Boxes []*Box
}

func (v *VisualSampleEntry) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &v.Entry)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &v.Extends); err != nil {
      return err
   }
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "avcC", "pasp", "sinf":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         v.Boxes = append(v.Boxes, &value)
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}

func (v VisualSampleEntry) Encode(w io.Writer) error {
   err := binary.Write(w, binary.BigEndian, v.Entry)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, v.Extends); err != nil {
      return err
   }
   for _, value := range v.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
