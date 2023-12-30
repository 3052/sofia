package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// Container: MovieFragmentBox
//  aligned(8) class TrackFragmentBox extends Box('traf') {
//  }
type TrackFragmentBox struct {
   Header BoxHeader
   Boxes  []Box
   TrackRun   TrackRunBox
   SampleEncryption   SampleEncryptionBox
}

func (b *TrackFragmentBox) Decode(r io.Reader) error {
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
      case "saio", "saiz", "sbgp", "sgpd", "tfdt", "tfhd":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := io.ReadFull(r, value.Payload)
         if err != nil {
            return err
         }
         b.Boxes = append(b.Boxes, value)
      case "trun":
         b.TrackRun.BoxHeader = head
         err := b.TrackRun.Decode(r)
         if err != nil {
            return err
         }
      case "senc":
         b.SampleEncryption.BoxHeader = head
         err := b.SampleEncryption.Decode(r)
         if err != nil {
            return err
         }
      case "uuid":
         decode := func() bool {
            if head.Extended_Type() == "a2394f525a9b4f14a2446c427c648df4" {
               if b.SampleEncryption.Sample_Count == 0 {
                  return true
               }
            }
            return false
         }
         if decode() {
            b.SampleEncryption.BoxHeader = head
            err := b.SampleEncryption.Decode(r)
            if err != nil {
               return err
            }
         } else {
            value := Box{Header: head}
            value.Payload = make([]byte, size)
            _, err := io.ReadFull(r, value.Payload)
            if err != nil {
               return err
            }
            b.Boxes = append(b.Boxes, value)
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (b TrackFragmentBox) Encode(w io.Writer) error {
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
   if err := b.TrackRun.Encode(w); err != nil {
      return err
   }
   return b.SampleEncryption.Encode(w)
}
