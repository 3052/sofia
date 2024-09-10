package file

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/mdat"
   "154.pages.dev/sofia/moof"
   "154.pages.dev/sofia/moov"
   "154.pages.dev/sofia/sidx"
   "io"
)

func (f *File) Write(dst io.Writer) error {
   // KEEP THESE IN ORDER
   for _, value := range f.Boxes {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.Movie != nil { // moov
      err := f.Movie.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.SegmentIndex != nil { // sidx
      err := f.SegmentIndex.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.MovieFragment != nil { // moof
      err := f.MovieFragment.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.MediaData != nil { // mdat
      err := f.MediaData.Write(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

// ISO/IEC 14496-12
type File struct {
   Boxes         []sofia.Box
   MediaData     *mdat.Box
   Movie         *moov.Box
   MovieFragment *moof.Box
   SegmentIndex  *sidx.Box
}

func (f *File) Read(r io.Reader) error {
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         _, size := head.GetSize()
         switch head.Type.String() {
         case "mdat":
            f.MediaData = &mdat.Box{}
            f.MediaData.Box.BoxHeader = head
            err := f.MediaData.Read(r)
            if err != nil {
               return err
            }
         case "moof":
            f.MovieFragment = &moof.Box{BoxHeader: head}
            err := f.MovieFragment.Read(r, size)
            if err != nil {
               return err
            }
         case "sidx":
            f.SegmentIndex = &sidx.Box{BoxHeader: head}
            err := f.SegmentIndex.Read(r)
            if err != nil {
               return err
            }
         case "moov":
            f.Movie = &moov.Box{BoxHeader: head}
            err := f.Movie.Read(r, size)
            if err != nil {
               return err
            }
         case "free", // Mubi
            "ftyp", // Roku
            "styp": // Roku
            object := sofia.Box{BoxHeader: head}
            err := object.Read(r)
            if err != nil {
               return err
            }
            f.Boxes = append(f.Boxes, object)
         default:
            var container sofia.Type
            copy(container[:], "File")
            return sofia.Error{container, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (f *File) GetMovie() (*moov.Box, bool) {
   if f.Movie != nil {
      return f.Movie, true
   }
   return nil, false
}
