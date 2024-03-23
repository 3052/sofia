package sofia

import (
   "errors"
   "io"
)

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
      "mvhd": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, &b)
      case "pssh":
         m.Protection.BoxHeader = head
         err := m.Protection.read(r)
         if err != nil {
            return err
         }
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

// ISO/IEC 14496-12
//  aligned(8) class MovieBox extends Box('moov') {
//  }
type Movie struct {
   BoxHeader BoxHeader
   Boxes     []*Box
   Protection ProtectionSystemSpecificHeader
   Track     Track
}

func (m Movie) write(w io.Writer) error {
   err := m.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   if err := m.Protection.write(w); err != nil {
      return err
   }
   return m.Track.write(w)
}
