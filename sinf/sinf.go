package sinf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/frma"
   "154.pages.dev/sofia/schi"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class ProtectionSchemeInfoBox(fmt) extends Box('sinf') {
//      OriginalFormatBox(fmt) original_format;
//      SchemeTypeBox scheme_type_box; // optional
//      SchemeInformationBox info; // optional
//   }
type Box struct {
   BoxHeader         sofia.BoxHeader
   Boxes             []sofia.Box
   OriginalFormat    frma.Box
   SchemeInformation schi.Box
}

func (b *Box) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "frma":
            b.OriginalFormat.BoxHeader = head
            err := b.OriginalFormat.Read(r)
            if err != nil {
               return err
            }
         case "schi":
            b.SchemeInformation.BoxHeader = head
            err := b.SchemeInformation.Read(r)
            if err != nil {
               return err
            }
         case "schm": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
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

func (b Box) write(w io.Writer) error {
   err := b.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   err = b.OriginalFormat.Write(w)
   if err != nil {
      return err
   }
   return b.SchemeInformation.Write(w)
}
