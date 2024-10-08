package container

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdat"
   "41.neocities.org/sofia/moof"
   "41.neocities.org/sofia/moov"
   "41.neocities.org/sofia/sidx"
)

// ISO/IEC 14496-12
type File struct {
   Box  []sofia.Box
   Mdat *mdat.Box
   Moof *moof.Box
   Moov *moov.Box
   Sidx *sidx.Box
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
      buf, err = f.Mdat.Box.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

func (f *File) Read(buf []byte) error {
   for len(buf) >= 1 {
      var value sofia.Box
      err := value.Read(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "free", // Mubi
         "ftyp", // Roku
         "styp": // Roku
         f.Box = append(f.Box, value)
      case "mdat":
         f.Mdat = &mdat.Box{value}
      case "moof":
         f.Moof = &moof.Box{BoxHeader: value.BoxHeader}
         err := f.Moof.Read(value.Payload)
         if err != nil {
            return err
         }
      case "moov":
         f.Moov = &moov.Box{BoxHeader: value.BoxHeader}
         err := f.Moov.Read(value.Payload)
         if err != nil {
            return err
         }
      case "sidx":
         f.Sidx = &sidx.Box{BoxHeader: value.BoxHeader}
         err := f.Sidx.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         var container sofia.BoxHeader
         copy(container.Type[:], "File")
         return &sofia.Error{container, value.BoxHeader}
      }
   }
   return nil
}

func (f *File) GetMoov() (*moov.Box, bool) {
   if f.Moov != nil {
      return f.Moov, true
   }
   return nil, false
}
