package sinf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/frma"
   "41.neocities.org/sofia/schi"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var boxVar sofia.Box
      err := boxVar.Read(data)
      if err != nil {
         return err
      }
      data = data[boxVar.BoxHeader.Size:]
      switch boxVar.BoxHeader.Type.String() {
      case "frma":
         b.Frma.BoxHeader = boxVar.BoxHeader
         err := b.Frma.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "schi":
         err := b.Schi.Read(boxVar.Payload)
         if err != nil {
            return err
         }
         b.Schi.BoxHeader = boxVar.BoxHeader
      case
         // amc-avc1
         // amc-mp4a
         // cineMember-avc1
         // criterion-avc1
         // criterion-mp4a
         // hboMax-dvh1
         // hboMax-ec-3
         // hboMax-hvc1
         // mubi-avc1
         // mubi-mp4a
         // nbc-avc1
         // nbc-mp4a
         // paramount-avc1
         // paramount-mp4a
         // plex-avc1
         // roku-avc1
         // roku-mp4a
         // tubi-avc1
         "schm":
         b.Box = append(b.Box, boxVar)
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
}

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

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   for _, boxVar := range b.Box {
      data, err = boxVar.Append(data)
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
