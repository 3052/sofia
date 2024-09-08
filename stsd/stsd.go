package stsd

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
   "io"
)

func (b *Box) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   err := b.FullBoxHeader.Read(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   for {
      var head sofia.BoxHeader
      err := head.Read(r)
      switch err {
      case nil:
         _, size := head.GetSize()
         switch head.Type.String() {
         case "enca":
            b.AudioSample = &AudioSampleEntry{}
            b.AudioSample.SampleEntry.BoxHeader = head
            err := b.AudioSample.read(r, size)
            if err != nil {
               return err
            }
         case "encv":
            b.VisualSample = &VisualSampleEntry{}
            b.VisualSample.SampleEntry.BoxHeader = head
            err := b.VisualSample.read(r, size)
            if err != nil {
               return err
            }
         case "avc1", // Tubi
            "ec-3", // Max
            "mp4a": // Tubi
            value := sofia.Box{BoxHeader: head}
            err := value.Read(r)
            if err != nil {
               return err
            }
            b.Boxes = append(b.Boxes, value)
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

func (b Box) SampleEntry() (*SampleEntry, bool) {
   if v := b.AudioSample; v != nil {
      return &v.SampleEntry, true
   }
   if v := b.VisualSample; v != nil {
      return &v.SampleEntry, true
   }
   return nil, false
}

func (b Box) write(w io.Writer) error {
   err := b.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, b.EntryCount)
   if err != nil {
      return err
   }
   for _, value := range b.Boxes {
      err := value.Write(w)
      if err != nil {
         return err
      }
   }
   if b.AudioSample != nil {
      err := b.AudioSample.write(w)
      if err != nil {
         return err
      }
   }
   if b.VisualSample != nil {
      err := b.VisualSample.write(w)
      if err != nil {
         return err
      }
   }
   return nil
}
func (b Box) Protection() (*sinf.Box, bool) {
   if v := b.AudioSample; v != nil {
      return &v.ProtectionScheme, true
   }
   if v := b.VisualSample; v != nil {
      return &v.ProtectionScheme, true
   }
   return nil, false
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
   Boxes         []sofia.Box
   AudioSample   *AudioSampleEntry
   VisualSample  *VisualSampleEntry
}
