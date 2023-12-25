package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MediaInformationBox extends Box('minf') {
// }
type MediaInformationBox struct {
   Header  BoxHeader
   Boxes []Box
   Stbl SampleTableBox
}

func (b *MediaInformationBox) Decode(r io.Reader) error {
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
      case "dinf", "smhd", "vmhd":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "stbl":
         b.Stbl.Header = head
         err := b.Stbl.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}

func (b MediaInformationBox) Encode(w io.Writer) error {
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
   return b.Stbl.Encode(w)
}
