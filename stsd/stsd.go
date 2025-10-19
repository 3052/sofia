package stsd

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/enca"
   "41.neocities.org/sofia/encv"
   "41.neocities.org/sofia/sinf"
   "encoding/binary"
)

func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
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
      var boxVar sofia.Box
      err := boxVar.Read(data)
      if err != nil {
         return err
      }
      data = data[boxVar.BoxHeader.Size:]
      switch boxVar.BoxHeader.Type.String() {
      case
         // rtbf-avc1
         // tubi-avc1
         "avc1",
         // hboMax-ec-3
         "ec-3":
         b.Box = append(b.Box, boxVar)
      case "enca":
         b.AudioSample = &enca.SampleEntry{}
         b.AudioSample.SampleEntry.BoxHeader = boxVar.BoxHeader
         err := b.AudioSample.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      case "encv":
         b.VisualSample = &encv.SampleEntry{}
         b.VisualSample.SampleEntry.BoxHeader = boxVar.BoxHeader
         err := b.VisualSample.Read(boxVar.Payload)
         if err != nil {
            return err
         }
      default:
         return &sofia.BoxError{b.BoxHeader, boxVar.BoxHeader}
      }
   }
   return nil
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, b.FullBoxHeader)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.EntryCount)
   for _, boxVar := range b.Box {
      data, err = boxVar.Append(data)
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
func (b *Box) Sinf() (*sinf.Box, bool) {
   if as := b.AudioSample; as != nil {
      return &as.Sinf, true
   }
   if vs := b.VisualSample; vs != nil {
      return &vs.Sinf, true
   }
   return nil, false
}

func (b *Box) SampleEntry() (*sofia.SampleEntry, bool) {
   if as := b.AudioSample; as != nil {
      return &as.SampleEntry, true
   }
   if vs := b.VisualSample; vs != nil {
      return &vs.SampleEntry, true
   }
   return nil, false
}

// ISO/IEC 14496-12
//
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
   Box           []sofia.Box
   AudioSample   *enca.SampleEntry
   VisualSample  *encv.SampleEntry
}
