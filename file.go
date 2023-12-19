package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragmentBox struct {
   Mfhd []byte
   Pssh []byte
   Traf []byte
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
      switch string(head.Type[:]) {
      case "mfhd":
         m.Mfhd = make([]byte, head.Size)
         _, err := r.Read(m.Mfhd)
         if err != nil {
            return err
         }
      case "pssh":
         pssh := make([]byte, head.Size)
         _, err := r.Read(pssh)
         if err != nil {
            return err
         }
         m.Pssh = append(m.Pssh, pssh...)
      case "traf":
         m.Traf = make([]byte, head.Size)
         _, err := r.Read(m.Traf)
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q\n", head.Type)
      }
   }
}
