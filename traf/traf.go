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
   for _, sb := range b.Box {
      buf, err = sb.Append(buf)
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
      var (
         sb sofia.Box
         err error
      )
      buf, err = sb.BoxHeader.Decode(buf)
      if err != nil {
         return err
      }
      switch sb.BoxHeader.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: sb.BoxHeader}
         buf, err = b.Senc.Decode(buf)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&sb.BoxHeader) {
            b.Senc = &senc.Box{BoxHeader: sb.BoxHeader}
            buf, err = b.Senc.Decode(buf)
            if err != nil {
               return err
            }
         } else {
            sb := sofia.Box{BoxHeader: sb.BoxHeader}
            buf, err = sb.Decode(buf)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, &sb)
         }
      case "saio", // Roku
      "saiz", // Roku
      "sbgp", // Roku
      "sgpd", // Roku
      "tfdt": // Roku
         sb := sofia.Box{BoxHeader: sb.BoxHeader}
         buf, err = sb.Decode(buf)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, &sb)
      case "tfhd":
         buf, err = b.Tfhd.Decode(buf)
         if err != nil {
            return err
         }
         b.Tfhd.BoxHeader = sb.BoxHeader
      case "trun":
         buf, err = b.Trun.Decode(buf)
         if err != nil {
            return err
         }
         b.Trun.BoxHeader = sb.BoxHeader
      default:
         return sofia.Error{b.BoxHeader.Type, sb.BoxHeader.Type}
      }
   }
   return nil
}
