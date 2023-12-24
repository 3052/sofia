package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieBox extends Box('moov') {
// }
type MovieBox struct {
   Header  BoxHeader
   Boxes []*Box
   Trak TrackBox
}

func (b MovieBox) Encode(w io.Writer) error {
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
   return b.Trak.Encode(w)
}

func (b *MovieBox) Decode(src io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(src)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "mvex", "mvhd", "pssh":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := src.Read(value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, &value)
      case "trak":
         b.Trak.Header = head
         err := b.Trak.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
