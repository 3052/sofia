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

///

func (b *Box) Decode(src []byte, size int64) ([]byte, error) {
   dst := src[size:]
   src = src[:size]
   for len(src) >= 1 {
      var head sofia.BoxHeader
      src, err = head.Decode(src)
      if err != nil {
         return nil, err
      }
      switch head.Type.String() {
      case "frma":
         src, err = b.Frma.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Frma.BoxHeader = head
      case "schi":
         src, err = b.Schi.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Schi.BoxHeader = head
      case "schm": // Roku
         value := sofia.Box{BoxHeader: head}
         src, err = value.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Box = append(b.Box, value)
      default:
         return nil, sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return dst, nil
}
