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
   return b.Mdia.Append(data)
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
      case "edts", // Paramount
         "tkhd", // Roku
         "tref", // RTBF
         "udta": // Mubi
         b.Box = append(b.Box, box0)
      case "mdia":
         b.Mdia.BoxHeader = box0.BoxHeader
         err := b.Mdia.Read(box0.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
      }
   }
   return nil
}
