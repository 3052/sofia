package sofia

import (
   "fmt"
   "io"
)

type File struct {
   Boxes []Box
   Moof  MovieFragmentBox
   Mdat  MediaDataBox
}

func (f File) Encode(dst io.Writer) error {
   for _, b := range f.Boxes {
      err := b.Encode(dst)
      if err != nil {
         return err
      }
   }
   err := f.Moof.Encode(dst)
   if err != nil {
      return err
   }
   return f.Mdat.Encode(dst)
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
      switch head.Type() {
      case "mdat":
         f.Mdat.Header = head
         err := f.Mdat.Decode(f.Moof.Traf.Trun, src)
         if err != nil {
            return err
         }
      case "moof":
         f.Moof.Header = head
         err := f.Moof.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      case "sidx", "styp":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         f.Boxes = append(f.Boxes, b)
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
