package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//  aligned(8) class MovieFragmentBox extends Box('moof') {
//  }
type MovieFragment struct {
   BoxHeader BoxHeader
   Boxes  []Box
   TrackFragment TrackFragment
}

func (m *MovieFragment) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      box_type := head.GetType()
      r := head.Payload(r)
      slog.Debug("BoxHeader", "Type", box_type)
      switch box_type {
      case "mfhd", // Roku
      "pssh": // Roku
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      case "traf":
         m.TrackFragment.BoxHeader = head
         err := m.TrackFragment.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("MovieFragment.Decode")
      }
   }
}

func (m MovieFragment) Encode(w io.Writer) error {
   err := m.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return m.TrackFragment.Encode(w)
}
