package sinf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/frma"
   "154.pages.dev/sofia/schi"
)

func (b *Box) Decode(buf []byte) error {
   for len(buf) >= 1 {
      var (
         sofia_box sofia.Box
         err error
      )
      buf, err = sofia_box.BoxHeader.Decode(buf)
      if err != nil {
         return err
      }
      buf = sofia_box.Decode(buf)
      switch head.Type.String() {
      case "frma":
         buf, err = b.Frma.Decode(buf)
         if err != nil {
            return err
         }
         b.Frma.BoxHeader = head
      case "schi":
         buf, err = b.Schi.Decode(buf)
         if err != nil {
            return err
         }
         b.Schi.BoxHeader = head
      case "schm": // Roku
         box_data := sofia.Box{BoxHeader: head}
         buf, err = box_data.Decode(buf)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, box_data)
      default:
         return sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return nil
}

// ISO/IEC 14496-12
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

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   for _, box_data := range b.Box {
      buf, err = box_data.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   buf, err = b.Frma.Append(buf)
   if err != nil {
      return nil, err
   }
   return b.Schi.Append(buf)
}
