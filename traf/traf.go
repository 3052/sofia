package traf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/senc"
   "41.neocities.org/sofia/tfhd"
   "41.neocities.org/sofia/trun"
   "log/slog"
)

func (b *Box) piff(head *sofia.BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if b.Senc == nil {
         return true
      }
   }
   return false
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", value.BoxHeader)
      data = data[value.BoxHeader.Size:]
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
   for _, value := range b.Box {
      data, err = value.Append(data)
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
