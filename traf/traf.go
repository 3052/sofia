package traf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/senc"
   "41.neocities.org/sofia/tfhd"
   "41.neocities.org/sofia/trun"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var boxVar sofia.Box
      err := boxVar.Read(data)
      if err != nil {
         return err
      }
      data = data[boxVar.BoxHeader.Size:]
      switch boxVar.BoxHeader.Type.String() {
      case "senc":
         b.Senc = &senc.Box{BoxHeader: boxVar.BoxHeader}
         err := b.Senc.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case
         // hulu-avc1
         "free",
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
         "saiz",
         // criterion-avc1
         // paramount-mp4a
         // roku-avc1
         // roku-mp4a
         "sbgp",
         // criterion-avc1
         // paramount-mp4a
         // roku-avc1
         // roku-mp4a
         "sgpd",
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
         // tubi-avc1
         "tfdt":
         b.Box = append(b.Box, &boxVar)
      case "tfhd":
         b.Tfhd.BoxHeader = boxVar.BoxHeader
         err := b.Tfhd.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "trun":
         b.Trun.BoxHeader = boxVar.BoxHeader
         err := b.Trun.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "uuid":
         if b.piff(&boxVar) {
            b.Senc = &senc.Box{BoxHeader: boxVar.BoxHeader}
            err := b.Senc.Read(boxVar.Payload)
            if err != nil {
               return err
            }
         } else {
            b.Box = append(b.Box, &boxVar)
         }
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
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
   for _, boxVar := range b.Box {
      data, err = boxVar.Append(data)
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

func (b *Box) piff(boxVar *sofia.Box) bool {
   if boxVar.BoxHeader.UserType.String() == sofia.PiffExtendedType {
      if b.Senc == nil {
         return true
      }
   }
   return false
}
