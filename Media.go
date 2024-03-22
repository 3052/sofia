package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MediaBox extends Box('mdia') {
//   }
type Media struct {
   BoxHeader        BoxHeader
   Boxes            []Box
   MediaInformation MediaInformation
}

func (m Media) write(w io.Writer) error {
   err := m.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return m.MediaInformation.write(w)
}

func (m *Media) read(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      box_type := head.GetType()
      slog.Debug("BoxHeader", "Type", box_type)
      ///////////////////////////////////////////////////////////////////////////
      r := head.payload(r)
      switch box_type {
      case "minf":
         m.MediaInformation.BoxHeader = head
         err := m.MediaInformation.read(r)
         if err != nil {
            return err
         }
      ///////////////////////////////////////////////////////////////////////////
      case "hdlr", // Roku
         "mdhd": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      default:
         return errors.New("Media.read")
      }
   }
}
