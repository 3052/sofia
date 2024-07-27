package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackBox extends Box('trak') {
//   }
type Track struct {
   BoxHeader BoxHeader
   Boxes     []Box
   Media     Media
}

func (t *Track) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.Read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
      case "mdia":
         _, size := head.get_size()
         t.Media.BoxHeader = head
         err := t.Media.read(r, size)
         if err != nil {
            return err
         }
      case "edts", // Paramount
      "tkhd", // Roku
      "tref", // RTBF
      "udta": // Mubi
         data := Box{BoxHeader: head}
         err := data.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, data)
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
   for _, data := range t.Boxes {
      err := data.write(w)
      if err != nil {
         return err
      }
   }
   return t.Media.write(w)
}
