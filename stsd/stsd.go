package stsd

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/enca"
   "154.pages.dev/sofia/encv"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
)

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.EntryCount)
   if err != nil {
      return nil, err
   }
   for _, box_data := range b.Box {
      buf, err = box_data.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if b.AudioSample != nil {
      buf, err = b.AudioSample.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   if b.VisualSample != nil {
      buf, err = b.VisualSample.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
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

func (b *Box) Decode(buf []byte, n int) error {
   buf, err := b.FullBoxHeader.Decode(buf[:n])
   if err != nil {
      return err
   }
   n, err = binary.Decode(buf, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   buf = buf[n:]
   for len(buf) >= 1 {
      var (
         head sofia.BoxHeader
         err error
      )
      buf, err = head.Decode(buf)
      if err != nil {
         return err
      }
      n = head.PayloadSize()
      switch head.Type.String() {
      case "enca":
         b.AudioSample = &enca.SampleEntry{}
         err := b.AudioSample.Decode(buf, n)
         if err != nil {
            return err
         }
         buf = buf[n:]
         b.AudioSample.SampleEntry.BoxHeader = head
      case "encv":
         b.VisualSample = &encv.SampleEntry{}
         err := b.VisualSample.Decode(buf, n)
         if err != nil {
            return err
         }
         buf = buf[n:]
         b.VisualSample.SampleEntry.BoxHeader = head
      case "avc1", // Tubi
      "ec-3", // Max
      "mp4a": // Tubi
         box_data := sofia.Box{BoxHeader: head}
         buf, err = box_data.Decode(buf)
         if err != nil {
            return err
         }
         b.Box = append(b.Box, box_data)
      default:
         return sofia.Error{b.BoxHeader.Type, head.Type}
      }
   }
   return nil
}
