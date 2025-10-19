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
      var box1 sofia.Box
      err := box1.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", box1.BoxHeader)
      data = data[box1.BoxHeader.Size:]
      switch box1.BoxHeader.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: box1.BoxHeader}
         err := b.Senc.Read(box1.Payload)
         if err != nil {
            return err
         }
      case
         // amc-avc1
         // amc-mp4a
         // cineMember-avc1
         // hboMax-dvh1
         // hboMax-ec-3
         // hboMax-hvc1
         // hulu-avc1
         // mubi-avc1
         // mubi-mp4a
         // nbc-avc1
         // nbc-mp4a
         // paramount-avc1
         // paramount-mp4a
         // plex-avc1
         // roku-avc1
         // roku-mp4a
         "saio",
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
         b.Box = append(b.Box, &box1)
      case "tfhd":
         b.Tfhd.BoxHeader = box1.BoxHeader
         err := b.Tfhd.Read(box1.Payload)
         if err != nil {
            return err
         }
      case "trun":
         b.Trun.BoxHeader = box1.BoxHeader
         err := b.Trun.Read(box1.Payload)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&box1) {
            b.Senc = &senc.Box{BoxHeader: box1.BoxHeader}
            err := b.Senc.Read(box1.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &box1)
         }
      default:
         return &sofia.BoxError{b.BoxHeader, box1.BoxHeader}
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
   for _, box1 := range b.Box {
      data, err = box1.Append(data)
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

func (b *Box) piff(box1 *sofia.Box) bool {
   if box1.BoxHeader.UserType.String() == sofia.PiffExtendedType {
      if b.Senc == nil {
         return true
      }
   }
   return false
}
