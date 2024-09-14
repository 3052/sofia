package moof

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
)

func (b *Box) Decode(buf []byte, size int64) error {
   buf = buf[:size]
   for len(buf) >= 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      buf, err = head.Decode(buf)
      if err != nil {
         return err
      }
      switch head.Type.String() {
      case "traf":
         n := head.PayloadSize()
         err := b.Traf.Decode(buf, n)
         if err != nil {
            return err
         }
         buf = buf[n:]
         b.Traf.BoxHeader = head
      case "mfhd", // Roku
      "pssh": // Roku
         box_data := sofia.Box{BoxHeader: head}
         buf, err = box_data.Decode(buf)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, box_data)
      default:
         return sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return nil
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   for _, box_data := range b.Box {
      buf, err = box_data.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Traf.Append(buf)
}

// ISO/IEC 14496-12
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type Box struct {
   BoxHeader sofia.BoxHeader
   Box       []sofia.Box
   Traf      traf.Box
}
