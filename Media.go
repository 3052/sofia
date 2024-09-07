package sofia

import (
   "154.pages.dev/sofia/box"
   "io"
)

func (m *Media) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "minf":
            _, size := head.get_size()
            m.MediaInformation.BoxHeader = head
            err := m.MediaInformation.read(r, size)
            if err != nil {
               return err
            }
         case "hdlr", // Roku
            "mdhd": // Roku
            object := box.Box{BoxHeader: head}
            err := object.read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, object)
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
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Media struct {
   BoxHeader        BoxHeader
   Boxes            []box.Box
   MediaInformation MediaInformation
}

func (m *Media) write(w io.Writer) error {
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
   return m.MediaInformation.write(w)
}
