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
         "sgpd", // Paramount
         "stco", // Roku
         "stsc", // Roku
         "stss", // CineMember
         "stsz", // Roku
         "stts": // Roku
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
