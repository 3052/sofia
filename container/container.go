package container

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/mdat"
   "154.pages.dev/sofia/moof"
   "154.pages.dev/sofia/moov"
   "154.pages.dev/sofia/sidx"
)

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

func (f *File) Append(buf []byte) ([]byte, error) {
   var err error
   // KEEP THESE IN ORDER
   for _, value := range f.Box {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if f.Moov != nil {
      buf, err = f.Moov.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if f.Sidx != nil {
      buf, err = f.Sidx.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if f.Moof != nil {
      buf, err = f.Moof.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if f.Mdat != nil {
      buf, err = f.Mdat.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

///

func (f *File) Read(src io.Reader) error {
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         size := head.PayloadSize()
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
