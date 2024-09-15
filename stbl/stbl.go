package stbl

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/stsd"
)

// ISO/IEC 14496-12
//   aligned(8) class SampleTableBox extends Box('stbl') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Stsd      stsd.Box
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   for _, value := range b.Box {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Stsd.Append(buf)
}

func (b *Box) Read(buf []byte) error {
   for len(buf) >= 1 {
      var value sofia.Box
      err := value.Read(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "sgpd", // Paramount
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
