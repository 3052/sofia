package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MediaBox extends Box('mdia') {
// }
type MediaBox struct {
   Header  BoxHeader
   Boxes []Box
   Minf MediaInformationBox
}

func (b *MediaBox) Decode(r io.Reader) error {
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
      case "hdlr", "mdhd":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "minf":
         b.Minf.Header = head
         err := b.Minf.Decode(r)
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}

func (b MediaBox) Encode(w io.Writer) error {
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
   return b.Minf.Encode(w)
}
