package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: MediaInformationBox
//  aligned(8) class SampleTableBox extends Box('stbl') {
//  }
type SampleTableBox struct {
   BoxHeader  BoxHeader
   Boxes []Box
   SampleDescription SampleDescriptionBox
}

func (s *SampleTableBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.BoxType())
      r := head.BoxPayload(r)
      switch head.BoxType() {
      case "sgpd", "stco", "stsc", "stsz", "stts":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         s.Boxes = append(s.Boxes, b)
      case "stsd":
         s.SampleDescription.BoxHeader = head
         err := s.SampleDescription.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("SampleTableBox.Decode")
      }
   }
}

func (s SampleTableBox) Encode(w io.Writer) error {
   err := s.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range s.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return s.SampleDescription.Encode(w)
}
