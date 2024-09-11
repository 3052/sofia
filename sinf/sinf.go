package sinf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/frma"
   "154.pages.dev/sofia/schi"
   "io"
)

// ISO/IEC 14496-12
//
//   aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//      OriginalFormatBox(fmt) original_format;
//      SchemeTypeBox scheme_type_box; // optional
//      SchemeInformationBox info; // optional
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Frma      frma.Box
   Schi      schi.Box
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "frma":
            b.Frma.BoxHeader = head
            err := b.Frma.Read(src)
            if err != nil {
               return err
            }
         case "schi":
            b.Schi.BoxHeader = head
            err := b.Schi.Read(src)
            if err != nil {
               return err
            }
         case "schm": // Roku
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
   err = b.Frma.Write(dst)
   if err != nil {
      return err
   }
   return b.Schi.Write(dst)
}
