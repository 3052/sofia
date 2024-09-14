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
   for _, value := range b.Box {
      buf, err = value.Append(buf)
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
      var value sofia.Box
      err := value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "frma":
         b.Frma.Decode(value.Payload)
         b.Frma.BoxHeader = value.BoxHeader
      case "schi":
         err := b.Schi.Decode(value.Payload)
         if err != nil {
            return err
         }
         b.Schi.BoxHeader = value.BoxHeader
      case "schm": // Roku
         b.Box = append(b.Box, value)
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
