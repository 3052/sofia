package minf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/stbl"
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
         // criterion-avc1
         // criterion-mp4a
         // hboMax-dvh1
         // hboMax-ec-3
         // hboMax-hvc1
         // mubi-avc1
         // mubi-mp4a
         // nbc-avc1
         // nbc-mp4a
         // paramount-avc1
         // paramount-mp4a
         // plex-avc1
         // roku-avc1
         // roku-mp4a
         // rtbf-avc1
         // tubi-avc1
         "dinf",
         // amc-mp4a
         // criterion-mp4a
         // hboMax-ec-3
         // mubi-mp4a
         // nbc-mp4a
         // paramount-mp4a
         // roku-mp4a
         "smhd",
         // amc-avc1
         // cineMember-avc1
         // criterion-avc1
         // hboMax-dvh1
         // hboMax-hvc1
         // mubi-avc1
         // nbc-avc1
         // paramount-avc1
         // plex-avc1
         // roku-avc1
         // rtbf-avc1
         // tubi-avc1
         "vmhd":
         b.Box = append(b.Box, boxVar)
      case "stbl":
         b.Stbl.BoxHeader = boxVar.BoxHeader
         err := b.Stbl.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
//   aligned(8) class MediaInformationBox extends Box('minf') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Stbl      stbl.Box
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
   return b.Stbl.Append(data)
}
