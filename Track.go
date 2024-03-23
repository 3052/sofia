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

func (t *Track) read(r io.Reader, size int64) error {
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
      case "mdia":
         _, size := head.get_size()
         t.Media.BoxHeader = head
         err := t.Media.read(r, size)
         if err != nil {
            return err
         }
      case "edts", // Paramount
         "tkhd", // Roku
         "udta": // Mubi
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      default:
         return errors.New("Track.read")
      }
   }
}
