package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MovieBox extends Box('moov') {
//   }
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
      box_type := head.GetType()
      slog.Debug("BoxHeader", "type", box_type)
      switch box_type {
      case "trak":
         _, size := head.get_size()
         m.Track.BoxHeader = head
         err := m.Track.read(r, size)
         if err != nil {
            return err
         }
      case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd", // Roku
         "pssh": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, &b)
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
   for _, b := range m.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return m.Track.write(w)
}
