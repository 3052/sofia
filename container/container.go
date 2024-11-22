package container

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdat"
   "41.neocities.org/sofia/moof"
   "41.neocities.org/sofia/moov"
   "41.neocities.org/sofia/sidx"
   "log/slog"
)

func (f *File) Read(data []byte) error {
   for len(data) >= 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      slog.Debug("box", "header", value.BoxHeader)
      data = data[value.BoxHeader.Size:]
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
   for _, value := range f.Box {
      data, err = value.Append(data)
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
      data, err = f.Mdat.Box.Append(data)
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
