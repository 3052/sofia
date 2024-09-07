package sofia

import (
   "154.pages.dev/sofia/box"
   "io"
)

func (m *MovieFragment) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head box.Header
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "traf":
            _, size := head.GetSize()
            m.TrackFragment.BoxHeader = head
            err := m.TrackFragment.read(r, size)
            if err != nil {
               return err
            }
         case "mfhd", // Roku
            "pssh": // Roku
            value := box.Box{BoxHeader: head}
            err := value.read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, value)
         default:
            return box.Error{m.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

// ISO/IEC 14496-12
//
//   aligned(8) class MovieFragmentBox extends Box('moof') {
//   }
type MovieFragment struct {
   BoxHeader     box.Header
   Boxes         []box.Box
   TrackFragment TrackFragment
}

func (m MovieFragment) write(w io.Writer) error {
   err := m.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   for _, value := range m.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   return m.TrackFragment.write(w)
}
