package trak

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/mdia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackBox extends Box('trak') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Mdia      mdia.Box
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   for _, sofia_box := range b.Box {
      buf, err = sofia_box.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Mdia.Append(buf)
}

func (b *Box) Decode(buf []byte) error {
   for len(buf) >= 1 {
      var sofia_box sofia.Box
      err := sofia_box.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[sofia_box.BoxHeader.Size:]
      switch sofix_box.BoxHeader.Type.String() {
      case "mdia":
         err := b.Mdia.Read(src, sofix_box.BoxHeader.PayloadSize())
         if err != nil {
            return err
         }
         b.Mdia.BoxHeader = sofix_box.BoxHeader
      case "edts", // Paramount
         "tkhd", // Roku
         "tref", // RTBF
         "udta": // Mubi
         sofia_box := sofia.Box{BoxHeader: sofix_box.BoxHeader}
         err := sofia_box.Read(src)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, sofia_box)
      default:
         return sofia.Error{b.BoxHeader.Type, sofix_box.BoxHeader.Type}
      }
   }
}
