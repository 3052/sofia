package stbl

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/stsd"
)

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      data = data[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "ctts", // FFmpeg
         "sgpd", // Paramount
         "stco", // Roku
         "stsc", // Roku
         "stss", // CineMember
         "stsz", // Roku
         "stts": // Roku
         b.Box = append(b.Box, value)
      case "stsd":
         b.Stsd.BoxHeader = value.BoxHeader
         err := b.Stsd.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
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
   for _, value := range b.Box {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Stsd.Append(data)
}
