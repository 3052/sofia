package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//  aligned(8) class SampleTableBox extends Box('stbl') {
//  }
type SampleTable struct {
   BoxHeader         BoxHeader
   Boxes             []Box
   SampleDescription SampleDescription
}

func (s *SampleTable) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
      case "stsd":
         _, size := head.get_size()
         s.SampleDescription.BoxHeader = head
         err := s.SampleDescription.read(r, size)
         if err != nil {
            return err
         }
      case "sgpd", // Paramount
      "stco", // Roku
      "stsc", // Roku
      "stsz", // Roku
      "stts": // Roku
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         s.Boxes = append(s.Boxes, object)
      default:
         return errors.New("SampleTable.read")
      }
   }
}

func (s SampleTable) write(w io.Writer) error {
   err := s.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, object := range s.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   return s.SampleDescription.write(w)
}
