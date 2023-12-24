package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class SampleTableBox extends Box('stbl') {
// }
type SampleTableBox struct {
   Header  BoxHeader
   Boxes []Box
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
      switch head.Type() {
      case "stco", "stsc", "stsd", "stsz", "stts":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      default:
         return fmt.Errorf("%q", head.RawType)
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
   return nil
}
