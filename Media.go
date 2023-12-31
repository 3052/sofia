package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: TrackBox
//  aligned(8) class MediaBox extends Box('mdia') {
//  }
type MediaBox struct {
   BoxHeader  BoxHeader
   Boxes []Box
   MediaInformation MediaInformationBox
}

func (m *MediaBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("*", "BoxType", head.BoxType())
      r := head.Reader(r)
      switch head.BoxType() {
      case "hdlr", "mdhd":
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
         return errors.New("BoxType")
      }
   }
}

func (m MediaBox) Encode(w io.Writer) error {
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
