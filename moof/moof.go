package moof

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
)

func (b *Box) Decode(buf []byte) error {
   for len(buf) >= 1 {
      var sof sofia.Box
      err := sof.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[sof.BoxHeader.Size:]
      switch sof.BoxHeader.Type.String() {
      case "traf":
         err := b.Traf.Decode(sof.Payload)
         if err != nil {
            return err
         }
         b.Traf.BoxHeader = sof.BoxHeader
      case "mfhd", // Roku
         "pssh": // Roku
         b.Box = append(b.Box, sof)
      default:
         return &sofia.Error{b.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
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
   return b.Traf.Append(buf)
}

// ISO/IEC 14496-12
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Traf      traf.Box
}
