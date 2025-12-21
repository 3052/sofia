package sofia

import (
   "encoding/binary"
   "errors"
   "io"
)

func (r *Remuxer) processFragment(moof *MoofBox, mdat *MdatBox) error {
   traf, ok := moof.Traf()
   if !ok {
      return nil
   }
   tfhd := traf.Tfhd()
   if tfhd == nil {
      return nil
   }
   // Senc is optional and provides encryption info
   senc, _ := traf.Senc()
   sencIndex := 0
   var newSamples []RemuxSample
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize
   defFlags := tfhd.DefaultSampleFlags
   mdatOffset := 0
   for _, child := range traf.Children {
      if child.Trun != nil {
         trun := child.Trun
         for i, sample := range trun.Samples {
            remuxSample := RemuxSample{Duration: defDur, Size: defSize, IsSync: true}
            // Determine flags logic:
            // 1. Start with TFHD Default Flags
            currentFlags := defFlags
            // 2. If 'FirstSampleFlags' is present (0x04) AND this is the first sample (index 0), use it.
            if i == 0 && (trun.Flags&0x000004) != 0 {
               currentFlags = trun.FirstSampleFlags
            }
            // 3. If 'SampleFlags' is present (0x400) in the array entry, it overrides everything.
            if (trun.Flags & 0x000400) != 0 {
               currentFlags = sample.Flags
            }
            if (trun.Flags & 0x000100) != 0 {
               remuxSample.Duration = sample.Duration
            }
            if (trun.Flags & 0x000200) != 0 {
               remuxSample.Size = sample.Size
            }
            // Check "sample_is_difference_sample" (bit 17, 0x10000)
            // 1 = Non-Sync (Difference)
            // 0 = Sync (Keyframe)
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
            var encInfo *SampleEncryptionInfo
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
   }
   if len(newSamples) == 0 {
      return nil
   }
   currentPos, _ := r.Writer.Seek(0, io.SeekCurrent)
   r.chunkOffsets = append(r.chunkOffsets, uint64(currentPos))
   if _, err := r.Writer.Write(mdat.Payload); err != nil {
      return err
   }
   r.samples = append(r.samples, newSamples...)
   r.segmentSampleCounts = append(r.segmentSampleCounts, uint32(len(newSamples)))
   return nil
}

func (r *Remuxer) Finish() error {
   if r.Moov == nil {
      return errors.New("not initialized")
   }
   mdatEndOffset, _ := r.Writer.Seek(0, io.SeekCurrent)
   finalMdatSize := uint64(mdatEndOffset - r.mdatStartOffset)
   var totalDuration uint64
   for _, sample := range r.samples {
      totalDuration += uint64(sample.Duration)
   }
   stts := buildStts(r.samples)
   stsz := buildStsz(r.samples)
   stsc := buildStsc(r.segmentSampleCounts)
   offsetBox := buildStco(r.chunkOffsets)
   stss := buildStss(r.samples)
   trak, _ := r.Moov.Trak()
   mdia, _ := trak.Mdia()
   minf, _ := mdia.Minf()
   stbl, _ := minf.Stbl()
   mdhd, ok := mdia.Mdhd()
   if !ok {
      return errors.New("missing mdhd")
   }
   // 1. Update Media Duration
   mdhd.SetDuration(totalDuration)
   // 2. Update Movie Header: Align Timescale and Duration
   if mvhd, ok := r.Moov.Mvhd(); ok {
      mvhd.Timescale = mdhd.Timescale // Simple logic: sync timescales
      mvhd.SetDuration(totalDuration)
   }
   r.Moov.RemoveMvex()
   trak.RemoveEdts()
   var newChildren []StblChild
   if stsd, ok := stbl.Stsd(); ok {
      stsd.UnprotectAll()
      newChildren = append(newChildren, StblChild{Stsd: stsd})
   } else {
      return errors.New("missing stsd")
   }
   newChildren = append(newChildren, StblChild{Raw: stts})
   newChildren = append(newChildren, StblChild{Raw: stsz})
   newChildren = append(newChildren, StblChild{Raw: stsc})
   newChildren = append(newChildren, StblChild{Raw: offsetBox})
   if stss != nil {
      newChildren = append(newChildren, StblChild{Raw: stss})
   }
   stbl.Children = newChildren
   moovBytes := r.Moov.Encode()
   if _, err := r.Writer.Write(moovBytes); err != nil {
      return err
   }
   if _, err := r.Writer.Seek(r.mdatStartOffset+8, io.SeekStart); err != nil {
      return err
   }
   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := r.Writer.Write(sizeBuf[:]); err != nil {
      return err
   }
   r.Writer.Seek(0, io.SeekEnd)
   return nil
}
