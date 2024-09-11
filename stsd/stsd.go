package stsd

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/enca"
   "154.pages.dev/sofia/encv"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

func (b *Box) Read(src io.Reader, size int64) error {
   src = io.LimitReader(src, size)
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   for {
      var head sofia.BoxHeader
      err := head.Read(src)
      switch err {
      case nil:
         _, size := head.GetSize()
         switch head.Type.String() {
         case "enca":
            b.AudioSample = &enca.SampleEntry{}
            b.AudioSample.SampleEntry.BoxHeader = head
            err := b.AudioSample.Read(src, size)
            if err != nil {
               return err
            }
         case "encv":
            b.VisualSample = &encv.SampleEntry{}
            b.VisualSample.SampleEntry.BoxHeader = head
            err := b.VisualSample.Read(src, size)
            if err != nil {
               return err
            }
         case "avc1", // Tubi
            "ec-3", // Max
            "mp4a": // Tubi
            value := sofia.Box{BoxHeader: head}
            err := value.Read(src)
            if err != nil {
               return err
            }
            b.Box = append(b.Box, value)
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

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.EntryCount)
   if err != nil {
      return err
   }
   for _, value := range b.Box {
      err := value.Write(dst)
      if err != nil {
         return err
      }
   }
   if b.AudioSample != nil {
      err := b.AudioSample.Write(dst)
      if err != nil {
         return err
      }
   }
   if b.VisualSample != nil {
      err := b.VisualSample.Write(dst)
      if err != nil {
         return err
      }
   }
   return nil
}

// ISO/IEC 14496-12
//   aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//      int i ;
//      unsigned int(32) entry_count;
//      for (i = 1 ; i <= entry_count ; i++){
//         SampleEntry(); // an instance of a class derived from SampleEntry
//      }
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   EntryCount    uint32
   Box         []sofia.Box
   AudioSample   *enca.SampleEntry
   VisualSample  *encv.SampleEntry
}

func (b *Box) SampleEntry() (*sofia.SampleEntry, bool) {
   if v := b.AudioSample; v != nil {
      return &v.SampleEntry, true
   }
   if v := b.VisualSample; v != nil {
      return &v.SampleEntry, true
   }
   return nil, false
}

func (b *Box) Sinf() (*sinf.Box, bool) {
   if v := b.AudioSample; v != nil {
      return &v.Sinf, true
   }
   if v := b.VisualSample; v != nil {
      return &v.Sinf, true
   }
   return nil, false
}
