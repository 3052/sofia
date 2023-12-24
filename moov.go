package sofia

import (
   "fmt"
   "io"
)

// aligned(8) class MovieBox extends Box('moov') {
// }
type MovieBox struct {
   Header  BoxHeader
   Boxes []Box
   Trak TrackBox
}

func (m MovieBox) Encode(dst io.Writer) error {
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
   return m.Trak.Encode(dst)
}

func (m *MovieBox) Decode(src io.Reader) error {
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
      case "mvex", "mvhd", "pssh":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      case "trak":
         m.Trak.Header = head
         err := m.Trak.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
