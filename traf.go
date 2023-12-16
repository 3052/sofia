package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragmentBox struct {
   Box Box
}

func (m *MovieFragmentBox) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &m.Box.Header.Size)
   if err != nil {
      return err
   }
   _, err = r.Read(m.Box.Header.Type[:])
   if err != nil {
      return err
   }
   m.Box.Payload = make([]byte, m.Box.Header.Size)
   _, err = r.Read(m.Box.Payload)
   if err != nil {
      return err
   }
   return nil
}
