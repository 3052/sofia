package sofia

import "errors"

// --- MOOF ---
type MoofBox struct {
   Header      BoxHeader
   Traf        *TrafBox
   Pssh        []*PsshBox
   RawChildren [][]byte
}

func (b *MoofBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      var header BoxHeader
      if err := header.Parse(payload[offset:]); err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "traf":
         var traf TrafBox
         if err := traf.Parse(content); err != nil {
            return err
         }
         b.Traf = &traf
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         b.Pssh = append(b.Pssh, &pssh)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

// --- TRAF ---
type TrafBox struct {
   Header      BoxHeader
   Tfhd        *TfhdBox
   Trun        []*TrunBox
   Senc        *SencBox
   RawChildren [][]byte
}

func (b *TrafBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      var header BoxHeader
      if err := header.Parse(payload[offset:]); err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "tfhd":
         var tfhd TfhdBox
         if err := tfhd.Parse(content); err != nil {
            return err
         }
         b.Tfhd = &tfhd
      case "trun":
         var trun TrunBox
         if err := trun.Parse(content); err != nil {
            return err
         }
         b.Trun = append(b.Trun, &trun)
      case "senc":
         var senc SencBox
         if err := senc.Parse(content); err != nil {
            return err
         }
         b.Senc = &senc
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

// --- TFHD ---
type TfhdBox struct {
   Header                 BoxHeader
   Flags                  uint32
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

func (b *TfhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 16 {
      return errors.New("tfhd too short")
   }
   p := parser{data: data, offset: 8}
   flags := p.Uint32()
   b.Flags = flags & 0x00FFFFFF
   b.TrackID = p.Uint32()

   if b.Flags&0x000001 != 0 { // base-data-offset-present
      if len(data) < p.offset+8 {
         return errors.New("tfhd too short for BaseDataOffset")
      }
      b.BaseDataOffset = p.Uint64()
   }
   if b.Flags&0x000002 != 0 { // sample-description-index-present
      if len(data) < p.offset+4 {
         return errors.New("tfhd too short for SampleDescriptionIndex")
      }
      b.SampleDescriptionIndex = p.Uint32()
   }
   if b.Flags&0x000008 != 0 { // default-sample-duration-present
      if len(data) < p.offset+4 {
         return errors.New("tfhd too short for DefaultSampleDuration")
      }
      b.DefaultSampleDuration = p.Uint32()
   }
   if b.Flags&0x000010 != 0 { // default-sample-size-present
      if len(data) < p.offset+4 {
         return errors.New("tfhd too short for DefaultSampleSize")
      }
      b.DefaultSampleSize = p.Uint32()
   }
   if b.Flags&0x000020 != 0 { // default-sample-flags-present
      if len(data) < p.offset+4 {
         return errors.New("tfhd too short for DefaultSampleFlags")
      }
      b.DefaultSampleFlags = p.Uint32()
   }
   return nil
}

// --- TRUN ---
type SampleInfo struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32
}

type TrunBox struct {
   Header           BoxHeader
   Flags            uint32
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Samples          []SampleInfo
}

func (b *TrunBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 16 {
      return errors.New("trun too short")
   }

   p := parser{data: data, offset: 8}
   flags := p.Uint32()
   b.Flags = flags & 0x00FFFFFF
   b.SampleCount = p.Uint32()

   if b.Flags&0x000001 != 0 {
      if len(data) < p.offset+4 {
         return errors.New("trun too short for data offset")
      }
      b.DataOffset = p.Int32()
   }
   if b.Flags&0x000004 != 0 {
      if len(data) < p.offset+4 {
         return errors.New("trun too short for first sample flags")
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
      return errors.New("trun box too short for declared samples")
   }

   b.Samples = make([]SampleInfo, b.SampleCount)
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
   return nil
}
