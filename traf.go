package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class TrackFragmentBox extends Box('traf') {
// }
type TrackFragmentBox struct {
   Header BoxHeader
   Senc SampleEncryptionBox
   Trun TrackRunBox
   Boxes []Box
}

func (t *TrackFragmentBox) Decode(src io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(src)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "senc":
         t.Senc.BoxHeader = head
         err := t.Senc.Decode(src)
         if err != nil {
            return err
         }
      case "trun":
         t.Trun.BoxHeader = head
         err := t.Trun.Decode(src)
         if err != nil {
            return err
         }
      case "saio", "saiz", "sbgp", "sgpd", "tfdt", "tfhd", "uuid":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
