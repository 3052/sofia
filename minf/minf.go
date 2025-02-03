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
   for _, box0 := range b.Box {
      data, err = box0.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Stbl.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box0 sofia.Box
      err := box0.Read(data)
      if err != nil {
         return err
      }
      data = data[box0.BoxHeader.Size:]
      switch box0.BoxHeader.Type.String() {
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
         b.Box = append(b.Box, box0)
      case "stbl":
         b.Stbl.BoxHeader = box0.BoxHeader
         err := b.Stbl.Read(box0.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
      }
   }
   return nil
}
