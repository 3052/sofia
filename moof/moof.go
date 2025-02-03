package moof

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/traf"
   "log/slog"
)

// ISO/IEC 14496-12
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Traf      traf.Box
}

func (b *Box) Read(data []byte) error {
   for len(data) >= 1 {
      var box0 sofia.Box
      err := box0.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", box0.BoxHeader)
      data = data[box0.BoxHeader.Size:]
      switch box0.BoxHeader.Type.String() {
      case "mfhd", // Roku
         "pssh": // Roku
         b.Box = append(b.Box, box0)
      case "traf":
         b.Traf.BoxHeader = box0.BoxHeader
         err := b.Traf.Read(box0.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, box0.BoxHeader}
      }
   }
   return nil
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
   return b.Traf.Append(data)
}
