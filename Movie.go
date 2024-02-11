package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: File
//  aligned(8) class MovieBox extends Box('moov') {
//  }
type MovieBox struct {
   BoxHeader BoxHeader
   Boxes []*Box
   Track TrackBox
}

func (m *MovieBox) Decode(r io.Reader) error {
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
      case "iods", "meta", "mvex", "mvhd", "pssh":
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
         return errors.New("BoxType")
      }
   }
}

func (m MovieBox) Encode(w io.Writer) error {
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
