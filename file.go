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

func (f *File) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "ftyp", "sidx", "styp":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         f.Boxes = append(f.Boxes, value)
      case "moov":
         f.Moov.Header = head
         err := f.Moov.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "moof":
         f.Moof.Header = head
         err := f.Moof.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "mdat":
         f.Mdat.Header = head
         err := f.Mdat.Decode(f.Moof.Traf.Trun, r)
         if err != nil {
            return err
         }
      default:
         return fmt.Errorf("file %q", head.RawType)
      }
   }
}

func (f File) Encode(w io.Writer) error {
   for _, value := range f.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Moov.Header.Size >= 1 {
      err := f.Moov.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Moof.Header.Size >= 1 {
      err := f.Moof.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Mdat.Header.Size >= 1 {
      err := f.Mdat.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
