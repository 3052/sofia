package stbl

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/stsd"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box1 sofia.Box
      err := box1.Read(data)
      if err != nil {
         return err
      }
      data = data[box1.BoxHeader.Size:]
      switch box1.BoxHeader.Type.String() {
      case "stsd":
         b.Stsd.BoxHeader = box1.BoxHeader
         err := b.Stsd.Read(box1.Payload)
         if err != nil {
            return err
         }
      case
         // criterion-avc1
         "sbgp",
         // criterion-avc1
         // paramount-mp4a
         "sgpd",
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
         "stco",
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
         "stsc",
         // cineMember-avc1
         "stss",
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
         "stsz",
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
         "stts":
         b.Box = append(b.Box, box1)
      default:
         return &sofia.BoxError{b.BoxHeader, box1.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
//
//   aligned(8) class SampleTableBox extends Box('stbl') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Stsd      stsd.Box
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
   return b.Stsd.Append(data)
}
