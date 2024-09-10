package traf

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/senc"
   "154.pages.dev/sofia/tfhd"
   "154.pages.dev/sofia/trun"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentBox extends Box('traf') {
//   }
type Box struct {
   BoxHeader        sofia.BoxHeader
   Boxes            []*sofia.Box
   FragmentHeader   tfhd.Box
   SampleEncryption *senc.Box
   TrackRun         trun.Box
}

func (b Box) piff(head sofia.BoxHeader) bool {
   if head.UserType.String() == "a2394f525a9b4f14a2446c427c648df4" {
      if b.SampleEncryption == nil {
         return true
      }
   }
   return false
}

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         switch head.Type.String() {
         case "senc":
            b.SampleEncryption = &senc.Box{BoxHeader: head}
            err := b.SampleEncryption.Read(src)
            if err != nil {
               return err
            }
         case "uuid":
            if b.piff(head) {
               b.SampleEncryption = &senc.Box{BoxHeader: head}
               err := b.SampleEncryption.Read(src)
               if err != nil {
                  return err
               }
            } else {
               value := sofia.Box{BoxHeader: head}
               err := value.Read(src)
               if err != nil {
                  return err
               }
               b.Boxes = append(b.Boxes, &value)
            }
         case "saio", // Roku
            "saiz", // Roku
            "sbgp", // Roku
            "sgpd", // Roku
            "tfdt": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            b.Boxes = append(b.Boxes, &value)
         case "tfhd":
            b.FragmentHeader.BoxHeader = head
            err := b.FragmentHeader.Read(src)
            if err != nil {
               return err
            }
         case "trun":
            b.TrackRun.BoxHeader = head
            err := b.TrackRun.Read(src)
            if err != nil {
               return err
            }
         default:
            return sofia.Error{b.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (b Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   err = b.FragmentHeader.Write(dst)
   if err != nil {
      return err
   }
   if b.SampleEncryption != nil {
      b.SampleEncryption.Write(dst)
   }
   return b.TrackRun.Write(dst)
}
