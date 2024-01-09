package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: MediaBox
//  aligned(8) class MediaInformationBox extends Box('minf') {
//  }
type MediaInformationBox struct {
   BoxHeader  BoxHeader
   Boxes []Box
   SampleTable SampleTableBox
}

func (m *MediaInformationBox) Decode(r io.Reader) error {
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
      case "dinf", "smhd", "vmhd":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      case "stbl":
         m.SampleTable.BoxHeader = head
         err := m.SampleTable.Decode(r)
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (m MediaInformationBox) Encode(w io.Writer) error {
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
   return m.SampleTable.Encode(w)
}
