package sofia

import (
   "fmt"
   "io"
)

// 8.5.1 Sample table box
//  aligned(8) class SampleTableBox extends Box('stbl') {
//  }
type SampleTableBox struct {
   Header  BoxHeader
   Boxes []Box
   Stsd SampleDescriptionBox
}

func (b *SampleTableBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.BoxType() {
      case "sgpd", "stco", "stsc", "stsz", "stts":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "stsd":
         b.Stsd.BoxHeader = head
         err := b.Stsd.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("stbl %q", head.Type)
      }
   }
}

func (b SampleTableBox) Encode(w io.Writer) error {
   err := b.Header.Encode(w)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return b.Stsd.Encode(w)
}
