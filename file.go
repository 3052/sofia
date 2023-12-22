package sofia

import (
   "fmt"
   "io"
)

type File struct {
   Moof MovieFragmentBox
   Boxes []Box
}

func (f *File) Decode(src io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(src)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.BoxType() {
      case "moof":
         f.Moof.Header = head
         err := f.Moof.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      case "mdat", "sidx", "styp":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         f.Boxes = append(f.Boxes, b)
      default:
         return fmt.Errorf("%q", head.Type)
      }
   }
}
