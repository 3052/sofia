package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class MovieFragmentHeaderBox extends FullBox('mfhd', 0, 0) {
//    unsigned int(32) sequence_number;
// }
type MovieFragmentHeader struct {
   BoxHeader BoxHeader
   FullBoxHeader FullBoxHeader
   Sequence_Number uint32
}

func (m *MovieFragmentHeader) Decode(r io.Reader) error {
   err := m.BoxHeader.Decode(r)
   if err != nil {
      return err
   }
   err = m.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   return binary.Read(r, binary.BigEndian, &m.Sequence_Number)
}
