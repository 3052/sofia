package mdia

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/minf"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Minf      minf.Box
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
   return b.Minf.Append(buf)
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
      case "hdlr", // Roku
         "mdhd": // Roku
         b.Box = append(b.Box, value)
      case "minf":
         b.Minf.BoxHeader = value.BoxHeader
         err := b.Minf.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
}
