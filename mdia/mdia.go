package mdia

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/minf"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Minf      minf.Box
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   for _, sof := range b.Box {
      buf, err = sof.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Minf.Append(buf)
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "minf":
            err := b.Minf.Read(src, head.PayloadSize())
            if err != nil {
               return err
            }
            b.Minf.BoxHeader = head
         case "hdlr", // Roku
            "mdhd": // Roku
            sof := sofia.Box{BoxHeader: head}
            err := sof.Read(src)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, sof)
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
