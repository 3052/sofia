package sofia

import (
   "fmt"
   "io"
)

type File struct {
   Moof MovieFragmentBox
   Mdat []byte
   Styp []byte
   Sidx []byte
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
      switch head.String() {
      case "mdat":
         io.CopyN(io.Discard, r, int64(head.Size)-8)
      case "moof":
         io.CopyN(io.Discard, r, int64(head.Size)-8)
      default:
         return fmt.Errorf("%q", head.Type)
      }
   }
}
