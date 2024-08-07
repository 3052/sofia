package sofia

import "io"

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type TrackFragment struct {
   BoxHeader        BoxHeader
   Boxes            []*Box
   FragmentHeader TrackFragmentHeader
   SampleEncryption *SampleEncryption
   TrackRun         TrackRun
}

func (t TrackFragment) piff(head BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if t.SampleEncryption == nil {
         return true
      }
   }
   return false
}

func (t *TrackFragment) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "senc":
            t.SampleEncryption = &SampleEncryption{BoxHeader: head}
            err := t.SampleEncryption.read(r)
            if err != nil {
               return err
            }
         case "uuid":
            if t.piff(head) {
               t.SampleEncryption = &SampleEncryption{BoxHeader: head}
               err := t.SampleEncryption.read(r)
               if err != nil {
                  return err
               }
            } else {
               data := Box{BoxHeader: head}
               err := data.read(r)
               if err != nil {
                  return err
               }
               t.Boxes = append(t.Boxes, &data)
            }
         case "saio", // Roku
         "saiz", // Roku
         "sbgp", // Roku
         "sgpd", // Roku
         "tfdt": // Roku
            data := Box{BoxHeader: head}
            err := data.read(r)
            if err != nil {
               return err
            }
            t.Boxes = append(t.Boxes, &data)
         case "tfhd":
            t.FragmentHeader.BoxHeader = head
            err := t.FragmentHeader.read(r)
            if err != nil {
               return err
            }
         case "trun":
            t.TrackRun.BoxHeader = head
            err := t.TrackRun.read(r)
            if err != nil {
               return err
            }
         default:
            return box_error{t.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (t TrackFragment) write(w io.Writer) error {
   err := t.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, data := range t.Boxes {
      err := data.write(w)
      if err != nil {
         return err
      }
   }
   err = t.FragmentHeader.write(w)
   if err != nil {
      return err
   }
   if t.SampleEncryption != nil {
      t.SampleEncryption.write(w)
   }
   return t.TrackRun.write(w)
}
