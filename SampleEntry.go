package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

func (a *AudioSampleEntry) Decode(r io.Reader) error {
   err := a.Entry.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &a.Extends); err != nil {
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
      size := head.BoxPayload()
      switch head.BoxType() {
      case "dec3", "esds":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         a.Boxes = append(a.Boxes, &value)
      case "sinf":
         a.ProtectionScheme.Header = head
         err := a.ProtectionScheme.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

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
   Extends struct {
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
   ProtectionScheme ProtectionSchemeInfoBox
}

// 8.5.2 Sample description box
//  aligned(8) abstract class SampleEntry(
//     unsigned int(32) format
//  ) extends Box(format) {
//     const unsigned int(8)[6] reserved = 0;
//     unsigned int(16) data_reference_index;
//  }
type SampleEntry struct {
   Header  BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
}

func (s *SampleEntry) Encode(w io.Writer) error {
   err := s.Header.Encode(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(s.Reserved[:]); err != nil {
      return err
   }
   return binary.Write(w, binary.BigEndian, s.Data_Reference_Index)
}

func (s *SampleEntry) Decode(r io.Reader) error {
   _, err := io.ReadFull(r, s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Read(r, binary.BigEndian, &s.Data_Reference_Index)
}

func (a AudioSampleEntry) Encode(w io.Writer) error {
   err := a.Entry.Encode(w)
   if err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, a.Extends); err != nil {
      return err
   }
   for _, value := range a.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return a.ProtectionScheme.Encode(w)
}

func (v *VisualSampleEntry) Decode(r io.Reader) error {
   err := v.Entry.Decode(r)
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
      slog.Debug("*", "BoxType", head.BoxType())
      size := head.BoxPayload()
      switch head.BoxType() {
      case "avcC", "pasp":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         v.Boxes = append(v.Boxes, &value)
      case "sinf":
         v.ProtectionScheme.Header = head
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
   if err := binary.Write(w, binary.BigEndian, v.Extends); err != nil {
      return err
   }
   for _, value := range v.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return v.ProtectionScheme.Encode(w)
}
