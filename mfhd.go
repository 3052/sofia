package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class MovieFragmentHeaderBox extends FullBox('mfhd', 0, 0) {
//    unsigned int(32) sequence_number;
// }
type MovieFragmentHeader struct {
   Box FullBox
   Sequence_Number uint32
}

func (m *MovieFragmentHeader) Decode(r io.Reader) error {
   err := m.Box.Decode(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, m.Sequence_Number)
   if err != nil {
      return err
   }
   return nil
}
