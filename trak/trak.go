package trak

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdia"
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
   for _, value := range b.Box {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Mdia.Append(buf)
}

func (b *Box) Read(buf []byte) error {
   for len(buf) >= 1 {
      var value sofia.Box
      err := value.Read(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "edts", // Paramount
         "tkhd", // Roku
         "tref", // RTBF
         "udta": // Mubi
         b.Box = append(b.Box, value)
      case "mdia":
         b.Mdia.BoxHeader = value.BoxHeader
         err := b.Mdia.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
