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
   for _, value := range b.Box {
      buf, err = value.Append(buf)
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

func (b *Box) Read(buf []byte) error {
   for len(buf) >= 1 {
      var value sofia.Box
      err := value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: value.BoxHeader}
         err := b.Senc.Decode(value.Payload)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&value.BoxHeader) {
            b.Senc = &senc.Box{BoxHeader: value.BoxHeader}
            err := b.Senc.Decode(value.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &value)
         }
      case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
         b.Box = append(b.Box, &value)
      case "tfhd":
         err := b.Tfhd.Decode(value.Payload)
         if err != nil {
            return err
         }
         b.Tfhd.BoxHeader = value.BoxHeader
      case "trun":
         err := b.Trun.Decode(value.Payload)
         if err != nil {
            return err
         }
         b.Trun.BoxHeader = value.BoxHeader
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
