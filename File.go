package sofia

import (
   "errors"
   "io"
)

func (f *File) Read(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         _, size := head.get_size()
         switch head.debug() {
         case "mdat":
            f.MediaData = new(MediaData)
            f.MediaData.Box.BoxHeader = head
            err := f.MediaData.read(r)
            if err != nil {
               return err
            }
         case "moof":
            f.MovieFragment = &MovieFragment{BoxHeader: head}
            err := f.MovieFragment.read(r, size)
            if err != nil {
               return err
            }
         case "sidx":
            f.SegmentIndex = &SegmentIndex{BoxHeader: head}
            err := f.SegmentIndex.read(r)
            if err != nil {
               return err
            }
         case "moov":
            f.Movie = &Movie{BoxHeader: head}
            err := f.Movie.read(r, size)
            if err != nil {
               return err
            }
         case "free", // Mubi
            "ftyp", // Roku
            "styp": // Roku
            object := Box{BoxHeader: head}
            err := object.read(r)
            if err != nil {
               return err
            }
            f.Boxes = append(f.Boxes, object)
         default:
            return errors.New("File.Read")
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

// ISO/IEC 14496-12
type File struct {
   Boxes         []Box
   MediaData     *MediaData
   Movie         *Movie
   MovieFragment *MovieFragment
   SegmentIndex  *SegmentIndex
}

func (f File) GetMovie() (*Movie, bool) {
   if f.Movie != nil {
      return f.Movie, true
   }
   return nil, false
}

func (f File) Write(w io.Writer) error {
   // KEEP THESE IN ORDER
   for _, value := range f.Boxes {
      err := value.write(w)
      if err != nil {
         return err
      }
   }
   if f.Movie != nil { // moov
      err := f.Movie.write(w)
      if err != nil {
         return err
      }
   }
   if f.SegmentIndex != nil { // sidx
      err := f.SegmentIndex.write(w)
      if err != nil {
         return err
      }
   }
   if f.MovieFragment != nil { // moof
      err := f.MovieFragment.write(w)
      if err != nil {
         return err
      }
   }
   if f.MediaData != nil { // mdat
      err := f.MediaData.write(w)
      if err != nil {
         return err
      }
   }
   return nil
}
