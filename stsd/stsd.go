package stsd

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/enca"
   "154.pages.dev/sofia/encv"
   "154.pages.dev/sofia/sinf"
   "encoding/binary"
)

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

func (b *Box) Read(buf []byte) error {
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
      var value sofia.Box
      err := value.Read(buf)
      if err != nil {
         return err
      }
      buf = buf[value.BoxHeader.Size:]
      switch value.BoxHeader.Type.String() {
      case "avc1", // Tubi
         "ec-3", // Max
         "mp4a": // Tubi
         b.Box = append(b.Box, value)
      case "enca":
         b.AudioSample = &enca.SampleEntry{}
         b.AudioSample.SampleEntry.BoxHeader = value.BoxHeader
         err := b.AudioSample.Read(value.Payload)
         if err != nil {
            return err
         }
      case "encv":
         b.VisualSample = &encv.SampleEntry{}
         b.VisualSample.SampleEntry.BoxHeader = value.BoxHeader
         err := b.VisualSample.Read(value.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.Error{b.BoxHeader, value.BoxHeader}
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

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf = binary.BigEndian.AppendUint32(buf, b.EntryCount)
   for _, value := range b.Box {
      buf, err = value.Append(buf)
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
