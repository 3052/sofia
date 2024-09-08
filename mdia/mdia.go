package mdia

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/minf"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader        sofia.BoxHeader
   Boxes            []sofia.Box
   MediaInformation minf.Box
}

func (b *Box) read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "minf":
            _, size := head.GetSize()
            b.MediaInformation.BoxHeader = head
            err := b.MediaInformation.read(src, size)
            if err != nil {
               return err
            }
         case "hdlr", // Roku
            "mdhd": // Roku
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

func (b *Box) write(dst io.Writer) error {
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
   return b.MediaInformation.write(dst)
}
