package sofia

import (
   "fmt"
   "io"
)

type File struct {
   Boxes []Box
   Moov MovieBox
   Moof MovieFragmentBox
   Mdat MediaDataBox
}

func (f File) Encode(dst io.Writer) error {
   for _, b := range f.Boxes {
      err := b.Encode(dst)
      if err != nil {
         return err
      }
   }
   if f.Moov.Header.Size >= 1 {
      err := f.Moov.Encode(dst)
      if err != nil {
         return err
      }
   }
   if f.Moof.Header.Size >= 1 {
      err := f.Moof.Encode(dst)
      if err != nil {
         return err
      }
   }
   if f.Mdat.Header.Size >= 1 {
      err := f.Mdat.Encode(dst)
      if err != nil {
         return err
      }
   }
   return nil
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
      case "ftyp", "sidx", "styp":
         b := Box{Header: head}
         b.Payload = make([]byte, size)
         _, err := src.Read(b.Payload)
         if err != nil {
            return err
         }
         f.Boxes = append(f.Boxes, b)
      case "moov":
         f.Moov.Header = head
         err := f.Moov.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      case "moof":
         f.Moof.Header = head
         err := f.Moof.Decode(io.LimitReader(src, size))
         if err != nil {
            return err
         }
      case "mdat":
         f.Mdat.Header = head
         err := f.Mdat.Decode(f.Moof.Traf.Trun, src)
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}
