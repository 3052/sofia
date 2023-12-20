package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragmentBox struct {
   Traf TrackFragmentBox
}

func (m *MovieFragmentBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.Size.Payload()
      switch head.Type.String() {
      case "traf":
         err := m.Traf.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "mfhd":
         io.CopyN(io.Discard, r, size)
      case "pssh":
         io.CopyN(io.Discard, r, size)
      default:
         return fmt.Errorf("%q", head.Type)
      }
   }
}
