package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

// Container: SampleDescriptionBox
//  class AudioSampleEntry(codingname) extends SampleEntry(codingname) {
//     const unsigned int(32)[2] reserved = 0;
//     unsigned int(16) channelcount;
//     template unsigned int(16) samplesize = 16;
//     unsigned int(16) pre_defined = 0;
//     const unsigned int(16) reserved = 0 ;
//     template unsigned int(32) samplerate = { default samplerate of media}<<16;
//  }
type AudioSampleEntry struct {
   Entry SampleEntry
   S struct {
      Reserved [2]uint32
      ChannelCount uint16
      SampleSize uint16
      Pre_Defined uint16
      _ uint16
      SampleRate uint32
   }
   Boxes []*Box
   ProtectionScheme ProtectionSchemeInfoBox
}

func (a *AudioSampleEntry) Decode(r io.Reader) error {
   err := a.Entry.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &a.S); err != nil {
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
      slog.Debug("*", "BoxType", head.BoxType())
      r := head.Reader(r)
      switch head.BoxType() {
      case "dec3", "esds":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         a.Boxes = append(a.Boxes, &b)
      case "sinf":
         a.ProtectionScheme.BoxHeader = head
         err := a.ProtectionScheme.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (a AudioSampleEntry) Encode(w io.Writer) error {
   err := a.Entry.Encode(w)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, a.S); err != nil {
      return err
   }
   for _, b := range a.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return a.ProtectionScheme.Encode(w)
}

// 8.5.2 Sample description box
//  aligned(8) abstract class SampleEntry(
//     unsigned int(32) format
//  ) extends Box(format) {
//     const unsigned int(8)[6] reserved = 0;
//     unsigned int(16) data_reference_index;
//  }
type SampleEntry struct {
   BoxHeader  BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
}

func (s *SampleEntry) Decode(r io.Reader) error {
   _, err := io.ReadFull(r, s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Read(r, binary.BigEndian, &s.Data_Reference_Index)
}

func (s *SampleEntry) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(s.Reserved[:]); err != nil {
      return err
   }
   return binary.Write(w, binary.BigEndian, s.Data_Reference_Index)
}

// Container: SampleDescriptionBox
//  class VisualSampleEntry(codingname) extends SampleEntry(codingname) {
//     unsigned int(16) pre_defined = 0;
//     const unsigned int(16) reserved = 0;
//     unsigned int(32)[3] pre_defined = 0;
//     unsigned int(16) width;
//     unsigned int(16) height;
//     template unsigned int(32) horizresolution = 0x00480000; // 72 dpi
//     template unsigned int(32) vertresolution = 0x00480000; // 72 dpi
//     const unsigned int(32) reserved = 0;
//     template unsigned int(16) frame_count = 1;
//     uint(8)[32] compressorname;
//     template unsigned int(16) depth = 0x0018;
//     int(16) pre_defined = -1;
//     // other boxes from derived specifications
//     CleanApertureBox clap; // optional
//     PixelAspectRatioBox pasp; // optional
//  }
type VisualSampleEntry struct {
   Entry SampleEntry
   S struct {
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
   ProtectionScheme ProtectionSchemeInfoBox
}

func (v *VisualSampleEntry) Decode(r io.Reader) error {
   err := v.Entry.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &v.S); err != nil {
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
      slog.Debug("*", "BoxType", head.BoxType())
      r := head.Reader(r)
      switch head.BoxType() {
      case "avcC", "pasp":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         v.Boxes = append(v.Boxes, &b)
      case "sinf":
         v.ProtectionScheme.BoxHeader = head
         err := v.ProtectionScheme.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (v VisualSampleEntry) Encode(w io.Writer) error {
   err := v.Entry.Encode(w)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, v.S); err != nil {
      return err
   }
   for _, b := range v.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return v.ProtectionScheme.Encode(w)
}
