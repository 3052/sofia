package sofia

import (
   "errors"
   "io"
   "log/slog"
)

// ISO/IEC 14496-12
//
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type TrackFragment struct {
   BoxHeader        BoxHeader
   Boxes            []Box
   SampleEncryption SampleEncryption
   TrackRun         TrackRun
}

func (t *TrackFragment) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      box_type := head.GetType()
      slog.Debug("BoxHeader", "type", box_type)
      switch box_type {
      case "uuid":
         decode := func() bool {
            if head.get_usertype() == "a2394f525a9b4f14a2446c427c648df4" {
               if t.SampleEncryption.SampleCount == 0 {
                  return true
               }
            }
            return false
         }
         if decode() {
            t.SampleEncryption.BoxHeader = head
            err := t.SampleEncryption.read(r)
            if err != nil {
               return err
            }
         } else {
            b := Box{BoxHeader: head}
            err := b.read(r)
            if err != nil {
               return err
            }
            t.Boxes = append(t.Boxes, b)
         }
      case "senc":
         t.SampleEncryption.BoxHeader = head
         err := t.SampleEncryption.read(r)
         if err != nil {
            return err
         }
      case "trun":
         t.TrackRun.BoxHeader = head
         err := t.TrackRun.read(r)
         if err != nil {
            return err
         }
      case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt", // Roku
         "tfhd": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, b)
      default:
         return errors.New("TrackFragment.read")
      }
   }
}

func (t TrackFragment) write(w io.Writer) error {
   err := t.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range t.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   if err := t.TrackRun.write(w); err != nil {
      return err
   }
   return t.SampleEncryption.write(w)
}
