package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: MovieBox
//  aligned(8) class TrackBox extends Box('trak') {
//  }
type TrackBox struct {
   BoxHeader  BoxHeader
   Boxes []Box
   Media MediaBox
}

func (b *TrackBox) Decode(r io.Reader) error {
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
      case "edts", "tkhd":
         value := Box{BoxHeader: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "mdia":
         b.Media.BoxHeader = head
         err := b.Media.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (b TrackBox) Encode(w io.Writer) error {
   err := b.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return b.Media.Encode(w)
}
