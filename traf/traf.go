package traf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/senc"
   "154.pages.dev/sofia/tfhd"
   "154.pages.dev/sofia/trun"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []*sofia.Box
   Tfhd      tfhd.Box
   Senc      *senc.Box
   Trun      trun.Box
}

func (b *Box) piff(head *sofia.BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if b.Senc == nil {
         return true
      }
   }
   return false
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
   buf, err = b.Tfhd.Append(buf)
   if err != nil {
      return nil, err
   }
   if b.Senc != nil {
      buf, err = b.Senc.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Trun.Append(buf)
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
      case "senc":
         b.Senc = &senc.Box{BoxHeader: sof.BoxHeader}
         err := b.Senc.Decode(buf)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&sof.BoxHeader) {
            b.Senc = &senc.Box{BoxHeader: sof.BoxHeader}
            err := b.Senc.Decode(sof.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &sof)
         }
      case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
         b.Box = append(b.Box, &sof)
      case "tfhd":
         err := b.Tfhd.Decode(sof.Payload)
         if err != nil {
            return err
         }
         b.Tfhd.BoxHeader = sof.BoxHeader
      case "trun":
         buf, err = b.Trun.Decode(buf)
         if err != nil {
            return err
         }
         b.Trun.BoxHeader = sof.BoxHeader
      default:
         return sofia.Error{b.BoxHeader.Type, sof.BoxHeader.Type}
      }
   }
   return nil
}
