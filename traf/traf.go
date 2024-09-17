package traf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/saio"
   "154.pages.dev/sofia/senc"
   "154.pages.dev/sofia/tfhd"
   "154.pages.dev/sofia/trun"
)

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.Tfhd.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.Tfdt.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.Trun.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.Saiz.Append(buf)
   if err != nil {
      return nil, err
   }
   if b.Saio != nil {
      buf, err = b.Saio.Append(buf)
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
   return b.Uuid.Append(buf)
   //for _, value := range b.Box {
   //   buf, err = value.Append(buf)
   //   if err != nil {
   //      return nil, err
   //   }
   //}
}

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
   Tfhd      tfhd.Box
   Tfdt sofia.Box
   Trun      trun.Box
   Saiz sofia.Box
   Saio      *saio.Box
   Senc      *senc.Box
   Uuid sofia.Box
   // Box       []*sofia.Box
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
      case "tfhd":
         b.Tfhd.BoxHeader = value.BoxHeader
         err := b.Tfhd.Read(value.Payload)
         if err != nil {
            return err
         }
      case "tfdt":
         b.Tfdt = value
      case "trun":
         b.Trun.BoxHeader = value.BoxHeader
         err := b.Trun.Read(value.Payload)
         if err != nil {
            return err
         }
      case "saiz":
         b.Saiz = value
      case "saio":
         b.Saio = &saio.Box{BoxHeader: value.BoxHeader}
         err := b.Saio.Read(value.Payload)
         if err != nil {
            return err
         }
      case "senc":
         b.Senc = &senc.Box{BoxHeader: value.BoxHeader}
         err := b.Senc.Read(value.Payload)
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
            //b.Box = append(b.Box, &value)
            b.Uuid = value
         }
      //case "saiz", // Roku
      //   "sbgp", // Roku
      //   "sgpd", // Roku
      //   "tfdt": // Roku
      //   b.Box = append(b.Box, &value)
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
