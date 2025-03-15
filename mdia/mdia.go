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
      var box1 sofia.Box
      err := box1.Read(data)
      if err != nil {
         return err
      }
      data = data[box1.BoxHeader.Size:]
      switch box1.BoxHeader.Type.String() {
      case "minf":
         b.Minf.BoxHeader = box1.BoxHeader
         err := b.Minf.Read(box1.Payload)
         if err != nil {
            return err
         }
      case "mdhd":
         b.Mdhd.BoxHeader = box1.BoxHeader
         err := b.Mdhd.Read(box1.Payload)
         if err != nil {
            return err
         }
      case "hdlr": // Roku
         b.Box = append(b.Box, box1)
      default:
         return &sofia.BoxError{b.BoxHeader, box1.BoxHeader}
      }
   }
   return nil
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
   data, err = b.Mdhd.Append(data)
   if err != nil {
      return nil, err
   }
   return b.Minf.Append(data)
}
