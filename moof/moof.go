package moof

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
)

func (b *Box) Decode(buf []byte) error {
   for len(buf) >= 1 {
      var value sofia.Box
      err := value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "traf":
         err := b.Traf.Decode(value.Payload)
         if err != nil {
            return err
         }
         b.Traf.BoxHeader = value.BoxHeader
      case "mfhd", // Roku
         "pssh": // Roku
         b.Box = append(b.Box, value)
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
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
