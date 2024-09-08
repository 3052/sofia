package stbl

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/stsd"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class SampleTableBox extends Box('stbl') {
//   }
type Box struct {
   BoxHeader         sofia.BoxHeader
   Boxes             []sofia.Box
   SampleDescription stsd.Box
}

func (s *Box) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "stsd":
            _, size := head.GetSize()
            s.SampleDescription.BoxHeader = head
            err := s.SampleDescription.Read(r, size)
            if err != nil {
               return err
            }
         case "sgpd", // Paramount
            "stco", // Roku
            "stsc", // Roku
            "stss", // CineMember
            "stsz", // Roku
            "stts": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            s.Boxes = append(s.Boxes, value)
         default:
            return sofia.Error{s.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (s Box) write(w io.Writer) error {
   err := s.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   for _, value := range s.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   return s.SampleDescription.Write(w)
}
