package mdia

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdhd"
   "41.neocities.org/sofia/minf"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Mdhd      mdhd.Box
   Minf      minf.Box
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
      case "minf":
         b.Minf.BoxHeader = boxVar.BoxHeader
         err := b.Minf.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "mdhd":
         b.Mdhd.BoxHeader = boxVar.BoxHeader
         err := b.Mdhd.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "hdlr": // Roku
         b.Box = append(b.Box, boxVar)
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
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
   data, err = b.Mdhd.Append(data)
   if err != nil {
      return nil, err
   }
   return b.Minf.Append(data)
}
