package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) abstract class SampleEntry(
//      unsigned int(32) format
//   ) extends Box(format) {
//      const unsigned int(8)[6] reserved = 0;
//      unsigned int(16) data_reference_index;
//   }
type SampleEntry struct {
   BoxHeader          BoxHeader
   Reserved           [6]uint8
   DataReferenceIndex uint16
}

func (s *SampleEntry) read(r io.Reader) error {
   _, err := io.ReadFull(r, s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Read(r, binary.BigEndian, &s.DataReferenceIndex)
}

func (s *SampleEntry) write(w io.Writer) error {
   err := s.BoxHeader.write(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(s.Reserved[:]); err != nil {
      return err
   }
   return binary.Write(w, binary.BigEndian, s.DataReferenceIndex)
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
type AudioSampleEntry struct {
   SampleEntry SampleEntry
   Extends     struct {
      _            [2]uint32
      ChannelCount uint16
      SampleSize   uint16
      PreDefined   uint16
      _            uint16
      SampleRate   uint32
   }
   Boxes            []*Box
   ProtectionScheme ProtectionSchemeInfo
}

func (v *VisualSampleEntry) read(r io.Reader) error {
   err := v.SampleEntry.read(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &v.Extends); err != nil {
      return err
   }
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.GetType())
      r := head.payload(r)
      switch head.GetType() {
      case "avcC", // Roku
         "btrt", // Mubi
         "colr", // Paramount
         "pasp": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         v.Boxes = append(v.Boxes, &b)
      case "sinf":
         v.ProtectionScheme.BoxHeader = head
         err := v.ProtectionScheme.read(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("VisualSampleEntry.Decode")
      }
   }
}

func (a *AudioSampleEntry) read(r io.Reader) error {
   err := a.SampleEntry.read(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &a.Extends); err != nil {
      return err
   }
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.GetType())
      r := head.payload(r)
      switch head.GetType() {
      case "dec3", // Hulu
         "esds": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         a.Boxes = append(a.Boxes, &b)
      case "sinf":
         a.ProtectionScheme.BoxHeader = head
         err := a.ProtectionScheme.read(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("AudioSampleEntry.Decode")
      }
   }
}

func (a AudioSampleEntry) write(w io.Writer) error {
   err := a.SampleEntry.write(w)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, a.Extends); err != nil {
      return err
   }
   for _, b := range a.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return a.ProtectionScheme.write(w)
}

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
   Boxes            []*Box
   ProtectionScheme ProtectionSchemeInfo
}

func (v VisualSampleEntry) write(w io.Writer) error {
   err := v.SampleEntry.write(w)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, v.Extends); err != nil {
      return err
   }
   for _, b := range v.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return v.ProtectionScheme.write(w)
}
