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

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "stsd":
            _, size := head.GetSize()
            b.SampleDescription.BoxHeader = head
            err := b.SampleDescription.Read(src, size)
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
            err := value.Read(src)
            if err != nil {
               return err
            }
            b.Boxes = append(b.Boxes, value)
         default:
            return sofia.Error{b.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (b Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   return b.SampleDescription.Write(dst)
}
