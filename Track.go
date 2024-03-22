package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class TrackBox extends Box('trak') {
//   }
type Track struct {
   BoxHeader BoxHeader
   Boxes     []Box
   Media     Media
}

func (t *Track) read(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.GetType())
      r := head.payload(r)
      switch head.GetType() {
      case "edts", // Paramount
         "tkhd", // Roku
         "udta": // Mubi
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      case "mdia":
         t.Media.BoxHeader = head
         err := t.Media.read(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("Track.read")
      }
   }
}

func (t Track) write(w io.Writer) error {
   err := t.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range t.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return t.Media.write(w)
}
