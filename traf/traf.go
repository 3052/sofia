package traf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/senc"
   "41.neocities.org/sofia/tfhd"
   "41.neocities.org/sofia/trun"
   "log/slog"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box0 sofia.Box
      err := box0.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", box0.BoxHeader)
      data = data[box0.BoxHeader.Size:]
      switch box0.BoxHeader.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: box0.BoxHeader}
         err := b.Senc.Read(box0.Payload)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&box0.BoxHeader) {
            b.Senc = &senc.Box{BoxHeader: box0.BoxHeader}
            err := b.Senc.Read(box0.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &box0)
         }
      case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
         b.Box = append(b.Box, &box0)
      case "tfhd":
         b.Tfhd.BoxHeader = box0.BoxHeader
         err := b.Tfhd.Read(box0.Payload)
         if err != nil {
            return err
         }
      case "trun":
         b.Trun.BoxHeader = box0.BoxHeader
         err := b.Trun.Read(box0.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
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
   Senc      *senc.Box
   Tfhd      tfhd.Box
   Trun      trun.Box
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   for _, box0 := range b.Box {
      data, err = box0.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if b.Senc != nil {
      data, err = b.Senc.Append(data)
      if err != nil {
         return nil, err
      }
   }
   data, err = b.Tfhd.Append(data)
   if err != nil {
      return nil, err
   }
   return b.Trun.Append(data)
}

func (b *Box) piff(head *sofia.BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if b.Senc == nil {
         return true
      }
   }
   return false
}
