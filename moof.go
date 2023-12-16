package sofia

import "io"

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragment struct {
   Mfhd MovieFragmentHeader
   Pssh []ProtectionSystemSpecificHeader
}

func (m *MovieFragment) Decode(r io.Reader) error {
   for {
      var b BoxHeader
      err := b.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch string(b.Type[:]) {
      case "mfhd":
         err := m.Mfhd.Decode(r)
         if err != nil {
            return err
         }
      case "pssh":
         var pssh ProtectionSystemSpecificHeader
         err := pssh.Decode(r)
         if err != nil {
            return err
         }
         m.Pssh = append(m.Pssh, pssh)
      }
   }
}
