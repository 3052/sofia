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
   for _, boxVar := range b.Box {
      data, err = boxVar.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return b.Stbl.Append(data)
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var boxVar sofia.Box
      err := boxVar.Read(data)
      if err != nil {
         return err
      }
      data = data[boxVar.BoxHeader.Size:]
      switch boxVar.BoxHeader.Type.String() {
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
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
