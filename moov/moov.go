package moov

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/pssh"
   "41.neocities.org/sofia/trak"
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
         // roku-avc1
         // roku-mp4a
         "iods",
         // paramount-avc1
         // paramount-mp4a
         // plex-avc1
         // tubi-avc1
         "meta",
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
         "mvex",
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
         "mvhd",
         // criterion-mp4a
         "udta":
         b.Box = append(b.Box, &boxVar)
      case "trak":
         b.Trak.BoxHeader = boxVar.BoxHeader
         err := b.Trak.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "pssh":
         pssh1 := pssh.Box{BoxHeader: boxVar.BoxHeader}
         err := pssh1.Read(boxVar.Payload)
         if err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, pssh1)
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []*sofia.Box
   Pssh      []pssh.Box
   Trak      trak.Box
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
   for _, boxVar := range b.Pssh {
      data, err = boxVar.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Trak.Append(data)
}
