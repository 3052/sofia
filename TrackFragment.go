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
   BoxHeader BoxHeader
   Boxes  []Box
   TrackRun   TrackRunBox
   SampleEncryption   SampleEncryptionBox
}

func (t *TrackFragmentBox) Decode(r io.Reader) error {
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      slog.Debug("*", "BoxType", head.BoxType())
      r := head.BoxPayload(r)
      switch head.BoxType() {
      case "saio", "saiz", "sbgp", "sgpd", "tfdt", "tfhd":
         b := Box{BoxHeader: head}
         err := b.Decode(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      case "trun":
         t.TrackRun.BoxHeader = head
         err := t.TrackRun.Decode(r)
         if err != nil {
            return err
         }
      case "senc":
         t.SampleEncryption.BoxHeader = head
         err := t.SampleEncryption.Decode(r)
         if err != nil {
            return err
         }
      case "uuid":
         decode := func() bool {
            if head.Extended_Type() == "a2394f525a9b4f14a2446c427c648df4" {
               if t.SampleEncryption.Sample_Count == 0 {
                  return true
               }
            }
            return false
         }
         if decode() {
            t.SampleEncryption.BoxHeader = head
            err := t.SampleEncryption.Decode(r)
            if err != nil {
               return err
            }
         } else {
            b := Box{BoxHeader: head}
            err := b.Decode(r)
            if err != nil {
               return err
            }
            t.Boxes = append(t.Boxes, b)
         }
      default:
         return errors.New("BoxType")
      }
   }
}

func (t TrackFragmentBox) Encode(w io.Writer) error {
   err := t.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   for _, b := range t.Boxes {
      err := b.Encode(w)
      if err != nil {
         return err
      }
   }
   if err := t.TrackRun.Encode(w); err != nil {
      return err
   }
   return t.SampleEncryption.Encode(w)
}
