package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Movie struct {
   BoxHeader  BoxHeader
   Boxes      []*Box
   Protection []ProtectionSystemSpecificHeader
   Track      Track
}

func (m *Movie) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.Read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
      case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd": // Roku
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, &object)
      case "pssh":
         protect := ProtectionSystemSpecificHeader{BoxHeader: head}
         err := protect.Read(r)
         if err != nil {
            return err
         }
         m.Protection = append(m.Protection, protect)
      case "trak":
         _, size := head.get_size()
         m.Track.BoxHeader = head
         err := m.Track.read(r, size)
         if err != nil {
            return err
         }
      default:
         return errors.New("Movie.read")
      }
   }
}

func (m Movie) write(w io.Writer) error {
   err := m.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, object := range m.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   for _, protect := range m.Protection {
      err := protect.Write(w)
      if err != nil {
         return err
      }
   }
   return m.Track.write(w)
}
