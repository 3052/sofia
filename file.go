package sofia

import (
   "errors"
   "io"
   "log/slog"
)

type File struct {
   Boxes []Box
   MediaData *MediaDataBox
   Movie *MovieBox
   MovieFragment *MovieFragmentBox
   SegmentIndex *SegmentIndexBox
}

// KEEP THESE IN ORDER
func (f File) Encode(w io.Writer) error {
   for _, value := range f.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.Movie != nil { // moov
      err := f.Movie.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.SegmentIndex != nil { // sidx
      err := f.SegmentIndex.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.MovieFragment != nil { // moof
      err := f.MovieFragment.Encode(w)
      if err != nil {
         return err
      }
   }
   if f.MediaData != nil { // mdat
      err := f.MediaData.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
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
         f.MediaData = new(MediaDataBox)
         f.MediaData.Header = head
         err := f.MediaData.Decode(
            f.MovieFragment.TrackFragment.TrackRun, io.LimitReader(r, size),
         )
         if err != nil {
            return err
         }
      case "moof":
         f.MovieFragment = new(MovieFragmentBox)
         f.MovieFragment.Header = head
         err := f.MovieFragment.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "moov":
         f.Movie = new(MovieBox)
         f.Movie.Header = head
         err := f.Movie.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      case "sidx":
         f.SegmentIndex = new(SegmentIndexBox)
         f.SegmentIndex.BoxHeader = head
         err := f.SegmentIndex.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}
