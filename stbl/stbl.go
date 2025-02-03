package stbl

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/stsd"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box0 sofia.Box
      err := box0.Read(data)
      if err != nil {
         return err
      }
      data = data[box0.BoxHeader.Size:]
      switch box0.BoxHeader.Type.String() {
      case "stsd":
         b.Stsd.BoxHeader = box0.BoxHeader
         err := b.Stsd.Read(box0.Payload)
         if err != nil {
            return err
         }
      case "ctts", // FFmpeg
         "sbgp", // Criterion
         "sgpd", // Paramount
         "stco", // Roku
         "stsc", // Roku
         "stss", // CineMember
         "stsz", // Roku
         "stts": // Roku
         b.Box = append(b.Box, box0)
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
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
   for _, box0 := range b.Box {
      data, err = box0.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Stsd.Append(data)
}
