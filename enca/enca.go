package enca

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

// ISO/IEC 14496-12
//   class AudioSampleEntry(codingname) extends SampleEntry(codingname) {
//      const unsigned int(32)[2] reserved = 0;
//      unsigned int(16) channelcount;
//      template unsigned int(16) samplesize = 16;
//      unsigned int(16) pre_defined = 0;
//      const unsigned int(16) reserved = 0 ;
//      template unsigned int(32) samplerate = { default samplerate of media}<<16;
//   }
type AudioSampleEntry struct {
   SampleEntry SampleEntry
   Extends     struct {
      _            [2]uint32
      ChannelCount uint16
      SampleSize   uint16
      PreDefined   uint16
      _            uint16
      SampleRate   uint32
   }
   Boxes            []*sofia.Box
   ProtectionScheme sinf.Box
}

func (a *AudioSampleEntry) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   err := a.SampleEntry.read(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &a.Extends)
   if err != nil {
      return err
   }
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         switch head.Type.String() {
         case "sinf":
            _, size := head.GetSize()
            a.ProtectionScheme.BoxHeader = head
            err := a.ProtectionScheme.Read(r, size)
            if err != nil {
               return err
            }
         case "dec3", // Hulu
            "esds": // Roku
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            a.Boxes = append(a.Boxes, &value)
         default:
            return sofia.Error{a.SampleEntry.BoxHeader.Type, head.Type}
         }
      case io.EOF:
         return nil
      default:
         return err
      }
   }
}

func (a AudioSampleEntry) write(w io.Writer) error {
   err := a.SampleEntry.write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, a.Extends)
   if err != nil {
      return err
   }
   for _, value := range a.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   return a.ProtectionScheme.Write(w)
}
