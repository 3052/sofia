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

func (b *MediaDataBox) Decode(r io.Reader, t TrackRunBox) error {
   b.Data = make([][]byte, t.Sample_Count)
   for i := range b.Data {
      var (
         data []byte
         err error
      )
      if size := t.Samples[i].Size; size >= 1 {
         data = make([]byte, size)
         _, err = io.ReadFull(r, data)
      } else {
         data, err = io.ReadAll(r)
      }
      if err != nil {
         return err
      }
      b.Data[i] = data
   }
   return nil
}

func (b MediaDataBox) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, data := range b.Data {
      _, err := w.Write(data)
      if err != nil {
         return err
      }
   }
   return nil
}
