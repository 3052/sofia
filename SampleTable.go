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

func (s SampleTable) write(w io.Writer) error {
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
   return s.SampleDescription.write(w)
}

func (s *SampleTable) read(r io.Reader, n int64) error {
   r = io.LimitReader(r, n)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      box_type := head.GetType()
      slog.Debug("BoxHeader", "type", box_type)
      switch box_type {
      case "stsd":
         s.SampleDescription.BoxHeader = head
         err := s.SampleDescription.read(r)
         if err != nil {
            return err
         }
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
      default:
         return errors.New("SampleTable.read")
      }
   }
}
