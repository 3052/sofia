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
   for _, value := range b.Box {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   for _, value := range b.Pssh {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Trak.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var sofia_box sofia.Box
      err := sofia_box.Read(data)
      if err != nil {
         return err
      }
      data = data[sofia_box.BoxHeader.Size:]
      switch sofia_box.BoxHeader.Type.String() {
      case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd", // Roku
         "udta": // Criterion
         b.Box = append(b.Box, &sofia_box)
      case "trak":
         b.Trak.BoxHeader = sofia_box.BoxHeader
         err := b.Trak.Read(sofia_box.Payload)
         if err != nil {
            return err
         }
      case "pssh":
         pssh_box := pssh.Box{BoxHeader: sofia_box.BoxHeader}
         err := pssh_box.Read(sofia_box.Payload)
         if err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, pssh_box)
      default:
         return &sofia.Error{b.BoxHeader, sofia_box.BoxHeader}
      }
   }
   return nil
}
