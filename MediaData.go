package sofia

import "io"

// Container: File
//  aligned(8) class MediaDataBox extends Box('mdat') {
//     bit(8) data[];
//  }
type MediaDataBox struct {
   BoxHeader BoxHeader
   Data [][]byte
}

func (m *MediaDataBox) Decode(r io.Reader, t TrackRunBox) error {
   m.Data = make([][]byte, t.SampleCount)
   for i := range m.Data {
      var err error
      if size := t.Sample[i].Size; size >= 1 {
         m.Data[i] = make([]byte, size)
         _, err = io.ReadFull(r, m.Data[i])
      } else {
         m.Data[i], err = io.ReadAll(r)
      }
      if err != nil {
         return err
      }
   }
   return nil
}

func (m MediaDataBox) Encode(w io.Writer) error {
   err := m.BoxHeader.Encode(w)
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
