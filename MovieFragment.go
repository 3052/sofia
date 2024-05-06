package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//  aligned(8) class MovieFragmentBox extends Box('moof') {
//  }
type MovieFragment struct {
   BoxHeader     BoxHeader
   Boxes         []Box
   TrackFragment TrackFragment
}

func (m *MovieFragment) read(r io.Reader, size int64) error {
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
      case "traf":
         _, size := head.get_size()
         m.TrackFragment.BoxHeader = head
         err := m.TrackFragment.read(r, size)
         if err != nil {
            return err
         }
      case "mfhd", // Roku
      "pssh": // Roku
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, object)
      default:
         return errors.New("MovieFragment.read")
      }
   }
}

func (m MovieFragment) write(w io.Writer) error {
   err := m.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, object := range m.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   return m.TrackFragment.write(w)
}
