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
   for _, value := range b.Box {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Stbl.Append(buf)
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
      case "stbl":
         b.Stbl.BoxHeader = value.BoxHeader
         err := b.Stbl.Decode(value.Payload)
         if err != nil {
            return err
         }
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
         b.Box = append(b.Box, value)
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
