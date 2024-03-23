package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type MovieFragment struct {
   BoxHeader     BoxHeader
   Boxes         []Box
   TrackFragment TrackFragment
}

func (m MovieFragment) write(w io.Writer) error {
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
   return m.TrackFragment.write(w)
}

func (m *MovieFragment) read(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      box_type := head.GetType()
      slog.Debug("BoxHeader", "Type", box_type)
      ///////////////////////////////////////////////////////////////////////////
      switch box_type {
      case "traf":
         m.TrackFragment.BoxHeader = head
         err := m.TrackFragment.read(r)
         if err != nil {
            return err
         }
      ///////////////////////////////////////////////////////////////////////////
      case "mfhd", // Roku
         "pssh": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      default:
         return errors.New("MovieFragment.read")
      }
   }
}
