package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: File
//  aligned(8) class MovieBox extends Box('moov') {
//  }
type MovieBox struct {
   Header BoxHeader
   Boxes []*Box
   Track TrackBox
}

func (b *MovieBox) Decode(r io.Reader) error {
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
      case "iods", "meta", "mvex", "mvhd", "pssh":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, &value)
      case "trak":
         b.Track.Header = head
         err := b.Track.Decode(io.LimitReader(r, size))
         if err != nil {
            return err
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (b MovieBox) Encode(w io.Writer) error {
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
   return b.Track.Encode(w)
}
