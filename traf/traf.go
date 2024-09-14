package traf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/senc"
   "154.pages.dev/sofia/tfhd"
   "154.pages.dev/sofia/trun"
)

func (b *Box) Decode(buf []byte, size int) error {
   buf = buf[:size]
   for len(buf) >= 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      buf, err = head.Decode(buf)
      if err != nil {
         return err
      }
      switch head.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: head}
         buf, err = b.Senc.Decode(buf)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&head) {
            b.Senc = &senc.Box{BoxHeader: head}
            buf, err = b.Senc.Decode(buf)
            if err != nil {
               return err
            }
         } else {
            box_data := sofia.Box{BoxHeader: head}
            buf, err = box_data.Decode(buf)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, &box_data)
         }
      case "saio", // Roku
      "saiz", // Roku
      "sbgp", // Roku
      "sgpd", // Roku
      "tfdt": // Roku
         box_data := sofia.Box{BoxHeader: head}
         buf, err = box_data.Decode(buf)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, &box_data)
      case "tfhd":
         buf, err = b.Tfhd.Decode(buf)
         if err != nil {
            return err
         }
         b.Tfhd.BoxHeader = head
      case "trun":
         buf, err = b.Trun.Decode(buf)
         if err != nil {
            return err
         }
         b.Trun.BoxHeader = head
      default:
         return sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return nil
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
   for _, box_data := range b.Box {
      buf, err = box_data.Append(buf)
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
