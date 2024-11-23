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
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", value.BoxHeader)
      data = data[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "mfhd", // Roku
         "pssh": // Roku
         b.Box = append(b.Box, value)
      case "traf":
         b.Traf.BoxHeader = value.BoxHeader
         err := b.Traf.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
      }
   }
   return nil
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
   return b.Traf.Append(data)
}
