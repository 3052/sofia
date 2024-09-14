package sinf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/frma"
   "154.pages.dev/sofia/schi"
)

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
   for _, sof := range b.Box {
      buf, err = sof.Append(buf)
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

func (b *Box) Decode(buf []byte) error {
   for len(buf) >= 1 {
      var sof sofia.Box
      err := sof.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[sof.BoxHeader.Size:]
      switch sof.BoxHeader.Type.String() {
      case "frma":
         b.Frma.Decode(sof.Payload)
         b.Frma.BoxHeader = sof.BoxHeader
      case "schi":
         err := b.Schi.Decode(sof.Payload)
         if err != nil {
            return err
         }
         b.Schi.BoxHeader = sof.BoxHeader
      case "schm": // Roku
         b.Box = append(b.Box, sof)
      default:
         return &sofia.Error{b.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
}
