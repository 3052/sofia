package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragmentBox struct {
   Header BoxHeader
   Boxes  []Box
   Traf   TrackFragmentBox
}

func (b *MovieFragmentBox) Decode(r io.Reader) error {
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
      case "traf":
         b.Traf.Header = head
         err := b.Traf.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "mfhd", "pssh":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      default:
         return fmt.Errorf("moof %q", head.RawType)
      }
   }
}

func (b MovieFragmentBox) Encode(w io.Writer) error {
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
   return b.Traf.Encode(w)
}
