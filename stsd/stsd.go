package stsd

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/enca"
   "41.neocities.org/sofia/encv"
   "41.neocities.org/sofia/sinf"
   "encoding/binary"
)

func (b *Box) Read(data []byte) error {
   n, err := b.FullBoxHeader.Decode(data)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.EntryCount)
   if err != nil {
      return err
   }
   data = data[n:]
   for len(data) >= 1 {
      var value sofia.Box
      err := value.Read(data)
      if err != nil {
         return err
      }
      data = data[value.BoxHeader.Size:]
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

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = b.FullBoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.EntryCount)
   for _, value := range b.Box {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if b.AudioSample != nil {
      data, err = b.AudioSample.Append(data)
      if err != nil {
         return nil, err
      }
   }
   if b.VisualSample != nil {
      data, err = b.VisualSample.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return data, nil
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
