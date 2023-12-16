package sofia

import "io"

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragment struct {
   BoxHeader BoxHeader
   Header MovieFragmentHeader
}

func (m *MovieFragment) Decode(r io.Reader) error {
   err := m.BoxHeader.Decode(r)
   if err != nil {
      return err
   }
   return m.Header.Decode(r)
}
