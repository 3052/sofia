package sofia

import "io"

// aligned(8) class TrackFragmentBox extends Box('traf') {
// }
type TrackFragmentBox struct {
   Tfhd []byte
}

func (t *TrackFragmentBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.String() {
      case "tfhd":
         t.Tfhd = make([]byte, head.Size)
         _, err := r.Read(t.Tfhd)
         if err != nil {
            return err
         }
      }
   }
}
