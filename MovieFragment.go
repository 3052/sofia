package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: File
//  aligned(8) class MovieFragmentBox extends Box('moof') {
//  }
type MovieFragmentBox struct {
   BoxHeader BoxHeader
   Boxes  []Box
   TrackFragment TrackFragmentBox
}

func (m *MovieFragmentBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.BoxType())
      r := head.BoxPayload(r)
      switch head.BoxType() {
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
         return errors.New("MovieFragmentBox.Decode")
      }
   }
}

func (m MovieFragmentBox) Encode(w io.Writer) error {
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
