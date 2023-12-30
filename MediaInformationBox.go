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
   Header  BoxHeader
   Boxes []Box
   SampleTable SampleTableBox
}

func (b *MediaInformationBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("*", "BoxType", head.BoxType())
      size := head.BoxPayload()
      switch head.BoxType() {
      case "dinf", "smhd", "vmhd":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "stbl":
         b.SampleTable.Header = head
         err := b.SampleTable.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (b MediaInformationBox) Encode(w io.Writer) error {
   err := b.Header.Encode(w)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return b.SampleTable.Encode(w)
}
