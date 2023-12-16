package sofia

import (
   "errors"
   "io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragment []Box

func (m *MovieFragment) Decode(r io.Reader) error {
   var b Box
   err := b.Decode(r)
   if err != nil {
      return err
   }
   *m, err = b.Boxes()
   if err != nil {
      return err
   }
   return nil
}

func (m MovieFragment) TrackFragment() (*TrackFragment, error) {
   for _, b := range m {
      if b.Type() == "traf" {
         return &TrackFragment{b}, nil
      }
   }
   return nil, errors.New("traf")
}
