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
   for _, box1 := range b.Box {
      data, err = box1.Append(data)
      if err != nil {
         return nil, err
      }
   }
   for _, box1 := range b.Pssh {
      data, err = box1.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Trak.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box1 sofia.Box
      err := box1.Read(data)
      if err != nil {
         return err
      }
      data = data[box1.BoxHeader.Size:]
      switch box1.BoxHeader.Type.String() {
      case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd", // Roku
         "udta": // Criterion
         b.Box = append(b.Box, &box1)
      case "trak":
         b.Trak.BoxHeader = box1.BoxHeader
         err := b.Trak.Read(box1.Payload)
         if err != nil {
            return err
         }
      case "pssh":
         pssh1 := pssh.Box{BoxHeader: box1.BoxHeader}
         err := pssh1.Read(box1.Payload)
         if err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, pssh1)
      default:
         return &sofia.BoxError{b.BoxHeader, box1.BoxHeader}
      }
   }
   return nil
}
