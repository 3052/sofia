package sinf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/frma"
   "41.neocities.org/sofia/schi"
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

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   for _, box0 := range b.Box {
      data, err = box0.Append(data)
      if err != nil {
         return nil, err
      }
   }
   data, err = b.Frma.Append(data)
   if err != nil {
      return nil, err
   }
   return b.Schi.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box0 sofia.Box
      err := box0.Read(data)
      if err != nil {
         return err
      }
      data = data[box0.BoxHeader.Size:]
      switch box0.BoxHeader.Type.String() {
      case "frma":
         b.Frma.BoxHeader = box0.BoxHeader
         err := b.Frma.Read(box0.Payload)
         if err != nil {
            return err
         }
      case "schi":
         err := b.Schi.Read(box0.Payload)
         if err != nil {
            return err
         }
         b.Schi.BoxHeader = box0.BoxHeader
      case "schm": // Roku
         b.Box = append(b.Box, box0)
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
      }
   }
   return nil
}
