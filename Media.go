package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//  aligned(8) class MediaBox extends Box('mdia') {
//  }
type Media struct {
   BoxHeader  BoxHeader
   Boxes []Box
   MediaInformation MediaInformation
}

func (m *Media) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("BoxHeader", "type", head.BoxType())
      r := head.BoxPayload(r)
      switch head.BoxType() {
      case "hdlr", // Roku
      "mdhd": // Roku
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      case "minf":
         m.MediaInformation.BoxHeader = head
         err := m.MediaInformation.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("Media.Decode")
      }
   }
}

func (m Media) Encode(w io.Writer) error {
   err := m.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   return m.MediaInformation.Encode(w)
}
