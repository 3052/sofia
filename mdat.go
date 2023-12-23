package sofia

import "io"

// aligned(8) class MediaDataBox extends Box('mdat') {
//    bit(8) data[];
// }
type MediaDataBox struct {
   Header BoxHeader
   Data [][]byte
}

func (m MediaDataBox) Encode(w io.Writer) error {
   err := m.Header.Encode(w)
   if err != nil {
      return err
   }
   for _, data := range m.Data {
      _, err := w.Write(data)
      if err != nil {
         return err
      }
   }
   return nil
}

func (m *MediaDataBox) Decode(t TrackRunBox, r io.Reader) error {
   for _, sample := range t.Samples {
      data := make([]byte, sample.Size)
      _, err := r.Read(data)
      if err != nil {
         return err
      }
      m.Data = append(m.Data, data)
   }
   return nil
}
