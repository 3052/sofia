package sofia

import (
   "errors"
   "fmt"
   "io"
)

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

// --- TRUN ---
type TrunSample struct {
   Size                  uint32
   Duration              uint32
   Flags                 uint32
   CompositionTimeOffset int32
}

// --- Logic ---
func (r *Remuxer) processFragment(moof *MoofBox, mdat *MdatBox) error {
   traf := moof.Traf
   if traf == nil {
      return nil
   }
   tfhd := traf.Tfhd
   if tfhd == nil {
      return nil
   }
   senc := traf.Senc
   sencIndex := 0
   var newSamples []*RemuxSample
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize
   defFlags := tfhd.DefaultSampleFlags
   mdatOffset := 0
   for _, trun := range traf.Trun {
      for i, sample := range trun.Samples {
         remuxSample := &RemuxSample{
            Duration:              defDur,
            Size:                  defSize,
            IsSync:                true,
            CompositionTimeOffset: 0,
         }
         currentFlags := defFlags
         // NOTE: The order of these two flag checks matters!
         // Per ISO/IEC 14496-12, if both sample_flags_present (0x000400) and
         // first_sample_flags_present (0x000004) are set, FirstSampleFlags
         // must OVERRIDE sample.Flags for the first sample (i==0).
         // Therefore, we must check sample_flags_present FIRST, then let
         // first_sample_flags_present overwrite it for i==0.
         // DO NOT swap these blocks, or FirstSampleFlags will be clobbered
         // by sample.Flags and the keyframe (sync sample) detection will be
         // corrupted for the first sample of each trun.
         if (trun.Flags & 0x000400) != 0 {
            currentFlags = sample.Flags
         }
         if i == 0 && (trun.Flags&0x000004) != 0 {
            currentFlags = trun.FirstSampleFlags
         }
         if (trun.Flags & 0x000100) != 0 {
            remuxSample.Duration = sample.Duration
         }
         if (trun.Flags & 0x000200) != 0 {
            remuxSample.Size = sample.Size
         }
         if (trun.Flags & 0x000800) != 0 {
            remuxSample.CompositionTimeOffset = sample.CompositionTimeOffset
         }
         if (currentFlags & 0x00010000) != 0 {
            remuxSample.IsSync = false
         } else {
            remuxSample.IsSync = true
         }
         originalSize := int(remuxSample.Size)
         if mdatOffset+originalSize > len(mdat.Payload) {
            return errors.New("mdat payload too short for samples")
         }
         sampleData := mdat.Payload[mdatOffset : mdatOffset+originalSize]
         var encInfo *SencSample
         if senc != nil && sencIndex < len(senc.Samples) {
            encInfo = &senc.Samples[sencIndex]
            sencIndex++
         }
         if r.OnSample != nil {
            r.OnSample(sampleData, encInfo)
         }
         newSamples = append(newSamples, remuxSample)
         mdatOffset += originalSize
      }
   }

   if len(newSamples) == 0 {
      return nil
   }
   currentPos, err := r.Writer.Seek(0, io.SeekCurrent)
   if err != nil {
      return fmt.Errorf("seeking to get chunk offset: %w", err)
   }
   r.chunkOffsets = append(r.chunkOffsets, uint64(currentPos))
   if _, err := r.Writer.Write(mdat.Payload); err != nil {
      return err
   }
   r.samples = append(r.samples, newSamples...)
   r.segmentSampleCounts = append(r.segmentSampleCounts, uint32(len(newSamples)))
   return nil
}

// fragment.go
