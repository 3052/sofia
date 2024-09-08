package file

import (
   "154.pages.dev/sofia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Box struct {
   BoxHeader        sofia.BoxHeader
   Boxes            []sofia.Box
   MediaInformation MediaInformation
}

func (m *Box) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "minf":
            _, size := head.GetSize()
            m.MediaInformation.BoxHeader = head
            err := m.MediaInformation.read(r, size)
            if err != nil {
               return err
            }
         case "hdlr", // Roku
            "mdhd": // Roku
            object := sofia.Box{BoxHeader: head}
            err := object.Read(r)
            if err != nil {
               return err
            }
            m.Boxes = append(m.Boxes, object)
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

func (m *Box) write(w io.Writer) error {
   err := m.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   for _, object := range m.Boxes {
      err := object.Write(w)
      if err != nil {
         return err
      }
   }
   return m.MediaInformation.write(w)
}
