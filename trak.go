package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class TrackBox extends Box('trak') {
// }
type TrackBox struct {
   Header  BoxHeader
   Boxes []Box
}

func (m TrackBox) Encode(dst io.Writer) error {
   err := m.Header.Encode(dst)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.Encode(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

func (m *TrackBox) Decode(src io.Reader) error {
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
      case "mvex", "mvhd", "pssh", "trak":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
