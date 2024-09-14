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

func (b *Box) Decode(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   buf = buf[n:]
   for len(buf) >= 1 {
      var sof sofia.Box
      err := sof.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[sof.BoxHeader.Size:]
      switch sof.BoxHeader.Type.String() {
      case "enca":
         b.AudioSample = &enca.SampleEntry{}
         b.AudioSample.SampleEntry.BoxHeader = sof.BoxHeader
         err := b.AudioSample.Decode(sof.Payload)
         if err != nil {
            return err
         }
      case "encv":
         b.VisualSample = &encv.SampleEntry{}
         b.VisualSample.SampleEntry.BoxHeader = sof.BoxHeader
         err := b.VisualSample.Decode(sof.Payload)
         if err != nil {
            return err
         }
      case "avc1", // Tubi
         "ec-3", // Max
         "mp4a": // Tubi
         b.Box = append(b.Box, sof)
      default:
         return &sofia.Error{b.BoxHeader, sof.BoxHeader}
      }
   }
   return nil
}
