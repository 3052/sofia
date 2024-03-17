package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//  aligned(8) class MovieBox extends Box('moov') {
//  }
type Movie struct {
   BoxHeader BoxHeader
   Boxes []*Box
   Track Track
}

func (m *Movie) Decode(r io.Reader) error {
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
      case "iods", // Roku
      "meta", // Paramount
      "mvex", // Roku
      "mvhd", // Roku
      "pssh": // Roku
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, &b)
      case "trak":
         m.Track.BoxHeader = head
         err := m.Track.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("Movie.Decode")
      }
   }
}

func (m Movie) Encode(w io.Writer) error {
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
   return m.Track.Encode(w)
}
