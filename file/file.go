package file

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdat"
   "41.neocities.org/sofia/moof"
   "41.neocities.org/sofia/moov"
   "41.neocities.org/sofia/sidx"
)

func (f *File) Read(data []byte) error {
   for len(data) >= 1 {
      var box sofia.Box
      err := box.Read(data)
      if err != nil {
         return err
      }
      data = data[box.BoxHeader.Size:]
      switch box.BoxHeader.Type.String() {
      case "free", // Mubi
         "ftyp", // Roku
         "styp": // Roku
         f.Box = append(f.Box, box)
      case "mdat":
         f.Mdat = &mdat.Box{box}
      case "moof":
         f.Moof = &moof.Box{BoxHeader: box.BoxHeader}
         err := f.Moof.Read(box.Payload)
         if err != nil {
            return err
         }
      case "moov":
         f.Moov = &moov.Box{BoxHeader: box.BoxHeader}
         err := f.Moov.Read(box.Payload)
         if err != nil {
            return err
         }
      case "sidx":
         f.Sidx = &sidx.Box{BoxHeader: box.BoxHeader}
         err := f.Sidx.Read(box.Payload)
         if err != nil {
            return err
         }
      default:
         var header sofia.BoxHeader
         copy(header.Type[:], "File")
         return &sofia.BoxError{header, box.BoxHeader}
      }
   }
   return nil
}

// ISO/IEC 14496-12
type File struct {
   Box  []sofia.Box
   Mdat *mdat.Box
   Moof *moof.Box
   Moov *moov.Box
   Sidx *sidx.Box
}

func (f *File) Append(data []byte) ([]byte, error) {
   var err error
   // KEEP THESE IN ORDER
   for _, box := range f.Box {
      data, err = box.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if f.Moov != nil {
      data, err = f.Moov.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if f.Sidx != nil {
      data, err = f.Sidx.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if f.Moof != nil {
      data, err = f.Moof.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if f.Mdat != nil {
      data, err = f.Mdat[0].Append(data)
      if err != nil {
         return nil, err
      }
   }
   return data, nil
}

func (f *File) GetMoov() (*moov.Box, bool) {
   if f.Moov != nil {
      return f.Moov, true
   }
   return nil, false
}
