package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//  aligned(8) class MovieBox extends Box('moov') {
//  }
type Movie struct {
   BoxHeader BoxHeader
   Boxes     []*Box
   Track     Track
}

func (m *Movie) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
      case "iods", // Roku
      "meta", // Paramount
      "mvex", // Roku
      "mvhd", // Roku
      "pssh": // Stan
         value := Box{BoxHeader: head}
         err := value.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, &value)
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
   for _, value := range m.Boxes {
      err := value.write(w)
      if err != nil {
         return err
      }
   }
   return m.Track.write(w)
}
