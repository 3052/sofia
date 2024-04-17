package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//  aligned(8) class TrackBox extends Box('trak') {
//  }
type Track struct {
   BoxHeader BoxHeader
   Boxes     []Box
   Media     Media
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
      "udta": // Mubi
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, object)
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
   for _, object := range t.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   return t.Media.write(w)
}
