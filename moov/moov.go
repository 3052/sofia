package moov

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/pssh"
   "154.pages.dev/sofia/trak"
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
   for _, pssh_box := range b.Pssh {
      buf, err = pssh_box.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Trak.Append(buf)
}

func (b *Box) Read(buf []byte) error {
   for len(buf) >= 1 {
      var sofia_box sofia.Box
      err := sofia_box.Read(buf)
      if err != nil {
         return err
      }
      buf = buf[sofia_box.BoxHeader.Size:]
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
