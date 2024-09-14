package minf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/stbl"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaInformationBox extends Box('minf') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Stbl      stbl.Box
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
   return b.Stbl.Append(buf)
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
      case "stbl":
         b.Stbl.BoxHeader = sof.BoxHeader
         err := b.Stbl.Decode(sof.Payload)
         if err != nil {
            return err
         }
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
         b.Box = append(b.Box, sof)
      default:
         return &sofia.Error{b.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
}
