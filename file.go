package sofia

import (
   "fmt"
   "io"
)

type File struct {
   Moof MovieFragmentBox
}

func (f *File) Decode(r io.Reader) error {
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
      case "moof":
         err := f.Moof.Decode(io.LimitReader(r, size))
         if err != nil {
            return fmt.Errorf("moof %v", err)
         }
      case "mdat":
         io.CopyN(io.Discard, r, size)
      case "sidx":
         io.CopyN(io.Discard, r, size)
      case "styp":
         io.CopyN(io.Discard, r, size)
      default:
         return fmt.Errorf("%q", head.Type)
      }
   }
}
