package traf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/senc"
   "41.neocities.org/sofia/tfhd"
   "41.neocities.org/sofia/trun"
)

func (b *Box) piff(head *sofia.BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if b.Senc == nil {
         return true
      }
   }
   return false
}

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []*sofia.Box
   Senc      *senc.Box
   Tfhd      tfhd.Box
   Trun      trun.Box
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
   if b.Senc != nil {
      buf, err = b.Senc.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   buf, err = b.Tfhd.Append(buf)
   if err != nil {
      return nil, err
   }
   return b.Trun.Append(buf)
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
      case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
         b.Box = append(b.Box, &value)
      case "senc":
         b.Senc = &senc.Box{BoxHeader: value.BoxHeader}
         err := b.Senc.Read(value.Payload)
         if err != nil {
            return err
         }
      case "tfhd":
         b.Tfhd.BoxHeader = value.BoxHeader
         err := b.Tfhd.Read(value.Payload)
         if err != nil {
            return err
         }
      case "trun":
         b.Trun.BoxHeader = value.BoxHeader
         err := b.Trun.Read(value.Payload)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&value.BoxHeader) {
            b.Senc = &senc.Box{BoxHeader: value.BoxHeader}
            err := b.Senc.Read(value.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &value)
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
