package traf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/senc"
   "154.pages.dev/sofia/tfhd"
   "154.pages.dev/sofia/trun"
)

func (b *Box) Decode(src []byte, size int64) ([]byte, error) {
   dst := src[size:]
   src = src[:size]
   for len(src) >= 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      src, err = head.Decode(src)
      if err != nil {
         return nil, err
      }
      switch head.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: head}
         src, err = b.Senc.Decode(src)
         if err != nil {
            return nil, err
         }
      case "uuid":
         if b.piff(&head) {
            b.Senc = &senc.Box{BoxHeader: head}
            src, err = b.Senc.Decode(src)
            if err != nil {
               return nil, err
            }
         } else {
            value := sofia.Box{BoxHeader: head}
            src, err = value.Decode(src)
            if err != nil {
               return nil, err
            }
            b.Box = append(b.Box, &value)
         }
      case "saio", // Roku
      "saiz", // Roku
      "sbgp", // Roku
      "sgpd", // Roku
      "tfdt": // Roku
         value := sofia.Box{BoxHeader: head}
         src, err = value.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Box = append(b.Box, &value)
      case "tfhd":
         src, err = b.Tfhd.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Tfhd.BoxHeader = head
      case "trun":
         src, err = b.Trun.Decode(src)
         if err != nil {
            return nil, err
         }
         b.Trun.BoxHeader = head
      default:
         return nil, sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return dst, nil
}

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
