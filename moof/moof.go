package moof

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   Boxes         []sofia.Box
   TrackFragment traf.Box
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "traf":
            _, size := head.GetSize()
            b.TrackFragment.BoxHeader = head
            err := b.TrackFragment.Read(src, size)
            if err != nil {
               return err
            }
         case "mfhd", // Roku
            "pssh": // Roku
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

func (b *Box) Write(dst io.Writer) error {
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
   return b.TrackFragment.Write(dst)
}
