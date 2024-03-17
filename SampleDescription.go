package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//  aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//     int i ;
//     unsigned int(32) entry_count;
//     for (i = 1 ; i <= entry_count ; i++){
//        SampleEntry(); // an instance of a class derived from SampleEntry
//     }
//  }
type SampleDescription struct {
   BoxHeader  BoxHeader
   FullBoxHeader FullBoxHeader
   EntryCount uint32
   AudioSample *AudioSampleEntry
   VisualSample *VisualSampleEntry
}

func (s *SampleDescription) Decode(r io.Reader) error {
   err := s.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.EntryCount); err != nil {
      return err
   }
   var head BoxHeader
   if err := head.Decode(r); err == io.EOF {
      return nil
   } else if err != nil {
      return err
   }
   slog.Debug("BoxHeader", "type", head.BoxType())
   switch head.BoxType() {
   case "enca":
      s.AudioSample = new(AudioSampleEntry)
      s.AudioSample.Entry.BoxHeader = head
      err := s.AudioSample.Decode(r)
      if err != nil {
         return err
      }
   case "encv":
      s.VisualSample = new(VisualSampleEntry)
      s.VisualSample.Entry.BoxHeader = head
      err := s.VisualSample.Decode(r)
      if err != nil {
         return err
      }
   default:
      return errors.New("SampleDescription.Decode")
   }
   return nil
}

func (s SampleDescription) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := s.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.EntryCount); err != nil {
      return err
   }
   if s.AudioSample != nil {
      err := s.AudioSample.Encode(w)
      if err != nil {
         return err
      }
   }
   if s.VisualSample != nil {
      err := s.VisualSample.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
