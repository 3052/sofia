package minf

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/stbl"
)

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
   for _, value := range b.Box {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Stbl.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      data = data[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
         b.Box = append(b.Box, value)
      case "stbl":
         b.Stbl.BoxHeader = value.BoxHeader
         err := b.Stbl.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
