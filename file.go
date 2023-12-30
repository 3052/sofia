package sofia

import (
   "errors"
   "io"
   "log/slog"
)

type File struct {
   Boxes []Box
   Movie MovieBox
   MovieFragment MovieFragmentBox
   Media MediaDataBox
   Segment SegmentIndexBox
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
      slog.Debug("*", "BoxType", head.BoxType())
      size := head.BoxPayload()
      switch head.BoxType() {
      case "ftyp", "styp":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         f.Boxes = append(f.Boxes, value)
      case "mdat":
         f.Media.Header = head
         err := f.Media.Decode(
            f.MovieFragment.Track.Trun, io.LimitReader(r, size),
         )
         if err != nil {
            return err
         }
      case "moof":
         f.MovieFragment.Header = head
         err := f.MovieFragment.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "moov":
         f.Movie.Header = head
         err := f.Movie.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "sidx":
         f.Segment.BoxHeader = head
         err := f.Segment.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
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
   if f.Movie.Header.Size >= 1 {
      err := f.Movie.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.MovieFragment.Header.Size >= 1 {
      err := f.MovieFragment.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Media.Header.Size >= 1 {
      err := f.Media.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Segment.BoxHeader.Size >= 1 {
      err := f.Segment.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
