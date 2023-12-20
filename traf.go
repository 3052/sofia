package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class TrackFragmentBox extends Box('traf') {
// }
type TrackFragmentBox struct {
   Senc SampleEncryptionBox
   Trun TrackRunBox
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
      size := head.Size.Payload()
      switch head.Type.String() {
      case "saio":
         io.CopyN(io.Discard, r, size)
      case "saiz":
         io.CopyN(io.Discard, r, size)
      case "sbgp":
         io.CopyN(io.Discard, r, size)
      case "senc":
         err := t.Senc.Decode(r)
         if err != nil {
            return err
         }
      case "sgpd":
         io.CopyN(io.Discard, r, size)
      case "tfdt":
         io.CopyN(io.Discard, r, size)
      case "tfhd":
         io.CopyN(io.Discard, r, size)
      case "trun":
         io.CopyN(io.Discard, r, size)
      default:
         return fmt.Errorf("%q", head.Type)
      }
   }
}
