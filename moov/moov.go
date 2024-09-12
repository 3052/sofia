package moov

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/pssh"
   "154.pages.dev/sofia/trak"
   "io"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []*sofia.Box
   Pssh      []pssh.Box
   Trak      trak.Box
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   for _, value := range b.Box {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   for _, protect := range b.Pssh {
      err := protect.Write(dst)
      if err != nil {
         return err
      }
   }
   return b.Trak.Write(dst)
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "iods", // Roku
            "meta", // Paramount
            "mvex", // Roku
            "mvhd", // Roku
            "udta": // Criterion
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, &value)
         case "pssh":
            protect := pssh.Box{BoxHeader: head}
            err := protect.Read(src)
            if err != nil {
               return err
            }
            b.Pssh = append(b.Pssh, protect)
         case "trak":
            _, size := head.GetSize()
            b.Trak.BoxHeader = head
            err := b.Trak.Read(src, size)
            if err != nil {
               return err
            }
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
