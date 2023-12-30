package sofia

import (
   "encoding/binary"
   "errors"
   "io"
   "log/slog"
)

// Container: SampleTableBox
//  aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//     int i ;
//     unsigned int(32) entry_count;
//     for (i = 1 ; i <= entry_count ; i++){
//        SampleEntry(); // an instance of a class derived from SampleEntry
//     }
//  }
type SampleDescriptionBox struct {
   BoxHeader  BoxHeader
   FullBoxHeader FullBoxHeader
   Entry_Count uint32
   AudioSample AudioSampleEntry
   VisualSample VisualSampleEntry
}

func (b *SampleDescriptionBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Entry_Count); err != nil {
      return err
   }
   var head BoxHeader
   if err := head.Decode(r); err == io.EOF {
      return nil
   } else if err != nil {
      return err
   }
   slog.Debug("*", "BoxType", head.BoxType())
   size := head.BoxPayload()
   switch head.BoxType() {
   case "enca":
      b.AudioSample.Header = head
      err := b.AudioSample.Decode(io.LimitReader(r, size))
      if err != nil {
         return err
      }
   case "encv":
      b.VisualSample.Header = head
      err := b.VisualSample.Decode(io.LimitReader(r, size))
      if err != nil {
         return err
      }
   default:
      return errors.New("BoxType")
   }
   return nil
}

func (s SampleDescriptionBox) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := s.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, s.Entry_Count); err != nil {
      return err
   }
   if s.AudioSample.Header.Size >= 1 {
      err := s.AudioSample.Encode(w)
      if err != nil {
         return err
      }
   }
   if s.VisualSample.Header.Size >= 1 {
      err := s.VisualSample.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
