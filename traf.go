package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class TrackFragmentBox extends Box('traf') {
// }
type TrackFragmentBox struct {
   Tfhd []byte
}

func (t *TrackFragmentBox) Decode(r io.Reader) error {
   for {
      var b BoxHeader
      err := b.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch string(b.Type[:]) {
      case "tfhd":
         t.Tfhd = make([]byte, b.Size)
         _, err := r.Read(t.Tfhd)
         if err != nil {
            return fmt.Errorf("tfhd %v", err)
         }
      }
   }
}
