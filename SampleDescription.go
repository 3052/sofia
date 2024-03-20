package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//      int i ;
//      unsigned int(32) entry_count;
//      for (i = 1 ; i <= entry_count ; i++){
//         SampleEntry(); // an instance of a class derived from SampleEntry
//      }
//   }
type SampleDescription struct {
   BoxHeader     BoxHeader
   FullBoxHeader FullBoxHeader
   EntryCount    uint32
   AudioSample   *AudioSampleEntry
   VisualSample  *VisualSampleEntry
}

func (s SampleDescription) write(w io.Writer) error {
   err := s.BoxHeader.write(w)
   if err != nil {
      return err
   }
   if err := s.FullBoxHeader.write(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.EntryCount); err != nil {
      return err
   }
   if s.AudioSample != nil {
      err := s.AudioSample.write(w)
      if err != nil {
         return err
      }
   }
   if s.VisualSample != nil {
      err := s.VisualSample.write(w)
      if err != nil {
         return err
      }
   }
   return nil
}

func (s *SampleDescription) read(r io.Reader) error {
   err := s.FullBoxHeader.read(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &s.EntryCount); err != nil {
      return err
   }
   var head BoxHeader
   if err := head.read(r); err == io.EOF {
      return nil
   } else if err != nil {
      return err
   }
   box_type := head.GetType()
   slog.Debug("BoxHeader", "Type", box_type)
   //////////////////////////////////////////////////////////////////////////////
   switch box_type {
   case "enca":
      s.AudioSample = new(AudioSampleEntry)
      s.AudioSample.SampleEntry.BoxHeader = head
      err := s.AudioSample.read(r)
      if err != nil {
         return err
      }
   case "encv":
      s.VisualSample = new(VisualSampleEntry)
      s.VisualSample.SampleEntry.BoxHeader = head
      err := s.VisualSample.read(r)
      if err != nil {
         return err
      }
   //////////////////////////////////////////////////////////////////////////////
   default:
      return errors.New("SampleDescription.Decode")
   }
   return nil
}
