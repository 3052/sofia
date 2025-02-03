package moov

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/pssh"
   "41.neocities.org/sofia/trak"
)

// ISO/IEC 14496-12
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []*sofia.Box
   Pssh      []pssh.Box
   Trak      trak.Box
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
   for _, box0 := range b.Pssh {
      data, err = box0.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Trak.Append(data)
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
      case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd", // Roku
         "udta": // Criterion
         b.Box = append(b.Box, &box0)
      case "trak":
         b.Trak.BoxHeader = box0.BoxHeader
         err := b.Trak.Read(box0.Payload)
         if err != nil {
            return err
         }
      case "pssh":
         pssh0 := pssh.Box{BoxHeader: box0.BoxHeader}
         err := pssh0.Read(box0.Payload)
         if err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, pssh0)
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
      }
   }
   return nil
}
