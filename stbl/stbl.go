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
   for _, box_data := range b.Box {
      buf, err = box_data.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return b.Stsd.Append(buf)
}

func (b *Box) Decode(buf []byte) ([]byte, error) {
   for len(buf) >= 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      buf, err = head.Decode(buf)
      if err != nil {
         return nil, err
      }
      var payload []byte
      payload, buf = head.Payload(buf)
      switch head.Type.String() {
      case "stsd":
         err := b.Stsd.Decode(payload)
         if err != nil {
            return err
         }
         b.Stsd.BoxHeader = head
      case "sgpd", // Paramount
         "stco", // Roku
         "stsc", // Roku
         "stss", // CineMember
         "stsz", // Roku
         "stts": // Roku
         box_data := sofia.Box{BoxHeader: head}
         err := box_data.Read(src)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, box_data)
      default:
         return sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
}
