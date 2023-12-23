package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieBox extends Box('moov') {
// }
type MovieBox struct {
   Header  BoxHeader
   Boxes []Box
}

func (MovieBox) Decode(src io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(src)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.Type() {
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}

func (MovieBox) Encode(io.Writer) error {
   return nil
}
