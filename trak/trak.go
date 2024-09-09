package trak

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/mdia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackBox extends Box('trak') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Boxes     []sofia.Box
   Media     mdia.Box
}

func (t *Box) Read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "mdia":
            _, size := head.GetSize()
            t.Media.BoxHeader = head
            err := t.Media.Read(r, size)
            if err != nil {
               return err
            }
         case "edts", // Paramount
            "tkhd", // Roku
            "tref", // RTBF
            "udta": // Mubi
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            t.Boxes = append(t.Boxes, value)
         default:
            return sofia.Error{t.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (t Box) Write(w io.Writer) error {
   err := t.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   for _, value := range t.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   return t.Media.Write(w)
}
