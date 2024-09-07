package sofia

import (
   "154.pages.dev/sofia/box"
   "io"
)

func (m *Movie) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head box.Header
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "iods", // Roku
         "meta", // Paramount
         "mvex", // Roku
         "mvhd": // Roku
            object := box.Box{BoxHeader: head}
            err := object.read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, &object)
         case "pssh":
            protect := ProtectionSystemSpecificHeader{BoxHeader: head}
            err := protect.Read(r)
            if err != nil {
               return err
            }
            m.Protection = append(m.Protection, protect)
         case "trak":
            _, size := head.get_size()
            m.Track.BoxHeader = head
            err := m.Track.read(r, size)
            if err != nil {
               return err
            }
         default:
            return box_error{m.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

// ISO/IEC 14496-12
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Movie struct {
   BoxHeader  box.Header
   Boxes      []*box.Box
   Protection []ProtectionSystemSpecificHeader
   Track      Track
}

func (m Movie) write(w io.Writer) error {
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
   for _, protect := range m.Protection {
      err := protect.Write(w)
      if err != nil {
         return err
      }
   }
   return m.Track.write(w)
}
