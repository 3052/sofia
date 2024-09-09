package file

import (
   "154.pages.dev/sofia"
   "io"
)

func (m *Movie) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "iods", // Roku
            "meta", // Paramount
            "mvex", // Roku
            "mvhd": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, &value)
         case "pssh":
            protect := ProtectionSystemSpecificHeader{BoxHeader: head}
            err := protect.Read(r)
            if err != nil {
               return err
            }
            m.Protection = append(m.Protection, protect)
         case "trak":
            _, size := head.GetSize()
            m.Track.BoxHeader = head
            err := m.Track.read(r, size)
            if err != nil {
               return err
            }
         default:
            return sofia.Error{m.BoxHeader.Type, head.Type}
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
//   aligned(8) class MovieBox extends Box('moov') {
//   }
type Movie struct {
   BoxHeader  sofia.BoxHeader
   Boxes      []*sofia.Box
   Protection []ProtectionSystemSpecificHeader
   Track      Track
}

func (m Movie) write(w io.Writer) error {
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
   for _, protect := range m.Protection {
      err := protect.Write(w)
      if err != nil {
         return err
      }
   }
   return m.Track.write(w)
}
