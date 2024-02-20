package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: MovieBox
//  aligned(8) class TrackBox extends Box('trak') {
//  }
type TrackBox struct {
   BoxHeader  BoxHeader
   Boxes []Box
   Media MediaBox
}

func (t *TrackBox) Decode(r io.Reader) error {
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
      case "edts", "tkhd":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      case "mdia":
         t.Media.BoxHeader = head
         err := t.Media.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("TrackBox.Decode")
      }
   }
}

func (t TrackBox) Encode(w io.Writer) error {
   err := t.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range t.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return t.Media.Encode(w)
}
