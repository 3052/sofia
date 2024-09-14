package mdia

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/minf"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Minf      minf.Box
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
   return b.Minf.Append(buf)
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
      case "minf":
         b.Minf.BoxHeader = sof.BoxHeader
         err := b.Minf.Decode(sof.Payload)
         if err != nil {
            return err
         }
      case "hdlr", // Roku
         "mdhd": // Roku
         b.Box = append(b.Box, sof)
      default:
         return &sofia.Error{b.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
}
