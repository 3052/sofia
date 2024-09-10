package moov

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/pssh"
   "154.pages.dev/sofia/trak"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Box struct {
   BoxHeader  sofia.BoxHeader
   Boxes      []*sofia.Box
   Protection []pssh.Box
   Track      trak.Box
}

func (b Box) write(dst io.Writer) error {
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
   for _, protect := range b.Protection {
      err := protect.Write(dst)
      if err != nil {
         return err
      }
   }
   return b.Track.Write(dst)
}

func (m *Box) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd", // Roku
         "udta": // Criterion
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, &value)
         case "pssh":
            protect := pssh.Box{BoxHeader: head}
            err := protect.Read(r)
            if err != nil {
               return err
            }
            m.Protection = append(m.Protection, protect)
         case "trak":
            _, size := head.GetSize()
            m.Track.BoxHeader = head
            err := m.Track.Read(r, size)
            if err != nil {
               return err
            }
         default:
            return sofia.Error{m.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}
