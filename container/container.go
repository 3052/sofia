package container

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
   for _, value := range f.Box {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.Moov != nil {
      err := f.Moov.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.Sidx != nil {
      err := f.Sidx.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.Moof != nil {
      err := f.Moof.Write(dst)
      if err != nil {
         return err
      }
   }
   if f.Mdat != nil {
      err := f.Mdat.Write(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

func (f *File) Read(src io.Reader) error {
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         _, size := head.GetSize()
         switch head.Type.String() {
         case "mdat":
            f.Mdat = &mdat.Box{}
            f.Mdat.Box.BoxHeader = head
            err := f.Mdat.Read(src)
            if err != nil {
               return err
            }
         case "moof":
            f.Moof = &moof.Box{BoxHeader: head}
            err := f.Moof.Read(src, size)
            if err != nil {
               return err
            }
         case "sidx":
            f.Sidx = &sidx.Box{BoxHeader: head}
            err := f.Sidx.Read(src)
            if err != nil {
               return err
            }
         case "moov":
            f.Moov = &moov.Box{BoxHeader: head}
            err := f.Moov.Read(src, size)
            if err != nil {
               return err
            }
         case "free", // Mubi
            "ftyp", // Roku
            "styp": // Roku
            object := sofia.Box{BoxHeader: head}
            err := object.Read(src)
            if err != nil {
               return err
            }
            f.Box = append(f.Box, object)
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

// ISO/IEC 14496-12
type File struct {
   Box  []sofia.Box
   Mdat *mdat.Box
   Moof *moof.Box
   Moov *moov.Box
   Sidx *sidx.Box
}

func (f *File) GetMoov() (*moov.Box, bool) {
   if f.Moov != nil {
      return f.Moov, true
   }
   return nil, false
}
