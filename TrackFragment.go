package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type TrackFragment struct {
   BoxHeader        BoxHeader
   Boxes            []*Box
   SampleEncryption *SampleEncryption
   TrackRun         TrackRun
}

func (t TrackFragment) write(w io.Writer) error {
   err := t.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, object := range t.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   if t.SampleEncryption != nil {
      t.SampleEncryption.write(w)
   }
   return t.TrackRun.write(w)
}

func (t *TrackFragment) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.Read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
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
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         t.Boxes = append(t.Boxes, &object)
      case "senc":
         t.SampleEncryption = &SampleEncryption{BoxHeader: head}
         err := t.SampleEncryption.read(r)
         if err != nil {
            return err
         }
      case "uuid":
         decode := func() bool {
            if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
               if t.SampleEncryption == nil {
                  return true
               }
            }
            return false
         }
         if decode() {
            t.SampleEncryption = &SampleEncryption{BoxHeader: head}
            err := t.SampleEncryption.read(r)
            if err != nil {
               return err
            }
         } else {
            object := Box{BoxHeader: head}
            err := object.read(r)
            if err != nil {
               return err
            }
            t.Boxes = append(t.Boxes, &object)
         }
      default:
         return errors.New("TrackFragment.read")
      }
   }
}
