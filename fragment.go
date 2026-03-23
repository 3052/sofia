// fragment.go
package sofia

import "errors"

// --- MOOF ---
type MoofBox struct {
   Header      *BoxHeader
   Traf        *TrafBox
   Pssh        []*PsshBox
   RawChildren [][]byte
}

func DecodeMoofBox(data []byte) (*MoofBox, error) {
   b := &MoofBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      header, err := DecodeBoxHeader(payload[offset:])
      if err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return nil, errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "traf":
         traf, err := DecodeTrafBox(content)
         if err != nil {
            return nil, err
         }
         b.Traf = traf
      case "pssh":
         pssh, err := DecodePsshBox(content)
         if err != nil {
            return nil, err
         }
         b.Pssh = append(b.Pssh, pssh)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

// --- TRAF ---
type TrafBox struct {
   Header      *BoxHeader
   Tfhd        *TfhdBox
   Trun        []*TrunBox
   Senc        *SencBox
   Tenc        *TencBox
   RawChildren [][]byte
}

func DecodeTrafBox(data []byte) (*TrafBox, error) {
   b := &TrafBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      header, err := DecodeBoxHeader(payload[offset:])
      if err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return nil, errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "tfhd":
         tfhd, err := DecodeTfhdBox(content)
         if err != nil {
            return nil, err
         }
         b.Tfhd = tfhd
      case "trun":
         trun, err := DecodeTrunBox(content)
         if err != nil {
            return nil, err
         }
         b.Trun = append(b.Trun, trun)
      case "senc":
         senc, err := DecodeSencBox(content)
         if err != nil {
            return nil, err
         }
         b.Senc = senc
      case "tenc":
         tenc, err := DecodeTencBox(content)
         if err != nil {
            return nil, err
         }
         b.Tenc = tenc
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

// --- TFHD ---
type TfhdBox struct {
   Header                 *BoxHeader
   Flags                  uint32
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

func DecodeTfhdBox(data []byte) (*TfhdBox, error) {
   b := &TfhdBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 16 {
      return nil, errors.New("tfhd too short")
   }
   p := parser{data: data, offset: 8}
   flags := p.Uint32()
   b.Flags = flags & 0x00FFFFFF
   b.TrackID = p.Uint32()

   if b.Flags&0x000001 != 0 { // base-data-offset-present
      if len(data) < p.offset+8 {
         return nil, errors.New("tfhd too short for BaseDataOffset")
      }
      b.BaseDataOffset = p.Uint64()
   }
   if b.Flags&0x000002 != 0 { // sample-description-index-present
      if len(data) < p.offset+4 {
         return nil, errors.New("tfhd too short for SampleDescriptionIndex")
      }
      b.SampleDescriptionIndex = p.Uint32()
   }
   if b.Flags&0x000008 != 0 { // default-sample-duration-present
      if len(data) < p.offset+4 {
         return nil, errors.New("tfhd too short for DefaultSampleDuration")
      }
      b.DefaultSampleDuration = p.Uint32()
   }
   if b.Flags&0x000010 != 0 { // default-sample-size-present
      if len(data) < p.offset+4 {
         return nil, errors.New("tfhd too short for DefaultSampleSize")
      }
      b.DefaultSampleSize = p.Uint32()
   }
   if b.Flags&0x000020 != 0 { // default-sample-flags-present
      if len(data) < p.offset+4 {
         return nil, errors.New("tfhd too short for DefaultSampleFlags")
      }
      b.DefaultSampleFlags = p.Uint32()
   }
   return b, nil
}

// --- TRUN ---
type TrunSample struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32
}

type TrunBox struct {
   Header           *BoxHeader
   Flags            uint32
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Samples          []TrunSample
}

func DecodeTrunBox(data []byte) (*TrunBox, error) {
   b := &TrunBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 16 {
      return nil, errors.New("trun too short")
   }

   p := parser{data: data, offset: 8}
   flags := p.Uint32()
   b.Flags = flags & 0x00FFFFFF
   b.SampleCount = p.Uint32()

   if b.Flags&0x000001 != 0 {
      if len(data) < p.offset+4 {
         return nil, errors.New("trun too short for data offset")
      }
      b.DataOffset = p.Int32()
   }
   if b.Flags&0x000004 != 0 {
      if len(data) < p.offset+4 {
         return nil, errors.New("trun too short for first sample flags")
      }
      b.FirstSampleFlags = p.Uint32()
   }

   sampleEntrySize := 0
   if b.Flags&0x000100 != 0 {
      sampleEntrySize += 4
   } // Duration
   if b.Flags&0x000200 != 0 {
      sampleEntrySize += 4
   } // Size
   if b.Flags&0x000400 != 0 {
      sampleEntrySize += 4
   } // Flags
   if b.Flags&0x000800 != 0 {
      sampleEntrySize += 4
   } // CTO
   if len(data)-p.offset < int(b.SampleCount)*sampleEntrySize {
      return nil, errors.New("trun box too short for declared samples")
   }

   b.Samples = make([]TrunSample, b.SampleCount)
   for i := uint32(0); i < b.SampleCount; i++ {
      if b.Flags&0x000100 != 0 {
         b.Samples[i].Duration = p.Uint32()
      }
      if b.Flags&0x000200 != 0 {
         b.Samples[i].Size = p.Uint32()
      }
      if b.Flags&0x000400 != 0 {
         b.Samples[i].Flags = p.Uint32()
      }
      if b.Flags&0x000800 != 0 {
         b.Samples[i].CompositionTimeOffset = p.Int32()
      }
   }
   return b, nil
}
