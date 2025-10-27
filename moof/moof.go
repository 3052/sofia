package moof

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/traf"
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
         // tubi-avc1
         "mfhd",
         // amc-avc1
         // amc-mp4a
         // cineMember-avc1
         // mubi-avc1
         // mubi-mp4a
         // nbc-avc1
         // nbc-mp4a
         // paramount-avc1
         // paramount-mp4a
         // plex-avc1
         // roku-avc1
         // roku-avc1
         // roku-mp4a
         // roku-mp4a
         // tubi-avc1
         "pssh":
         b.Box = append(b.Box, boxVar)
      case "traf":
         b.Traf.BoxHeader = boxVar.BoxHeader
         err := b.Traf.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
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
   return b.Traf.Append(data)
}

// ISO/IEC 14496-12
//
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Traf      traf.Box
}
