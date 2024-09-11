package minf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/stbl"
   "io"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MediaInformationBox extends Box('minf') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Stbl      stbl.Box
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "stbl":
            _, size := head.GetSize()
            b.Stbl.BoxHeader = head
            err := b.Stbl.Read(src, size)
            if err != nil {
               return err
            }
         case "dinf", // Roku
            "smhd", // Roku
            "vmhd": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, value)
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
   for _, value := range b.Box {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   return b.Stbl.Write(dst)
}
