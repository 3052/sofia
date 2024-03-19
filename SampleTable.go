package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class SampleTableBox extends Box('stbl') {
//   }
type SampleTable struct {
   BoxHeader         BoxHeader
   Boxes             []Box
   SampleDescription SampleDescription
}

func (s *SampleTable) Decode(r io.Reader) error {
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
      case "sgpd", // Paramount
         "stco", // Roku
         "stsc", // Roku
         "stsz", // Roku
         "stts": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
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
         return errors.New("SampleTable.Decode")
      }
   }
}

func (s SampleTable) Encode(w io.Writer) error {
   err := s.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range s.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return s.SampleDescription.Encode(w)
}
