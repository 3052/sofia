package sofia

import (
   "errors"
   "fmt"
)

// buildChunkOffsetBox decides whether to use stco or co64.
func buildChunkOffsetBox(offsets []uint64) []byte {
   for _, offset := range offsets {
      if offset > 0xFFFFFFFF {
         return encodeTable64("co64", offsets)
      }
   }
   entries32 := make([]uint32, len(offsets))
   for i, offset := range offsets {
      entries32[i] = uint32(offset)
   }
   return encodeTable("stco", len(entries32), entries32)
}

func buildCtts(samples []*RemuxSample) []byte {
   hasCTO := false
   for _, sample := range samples {
      if sample.CompositionTimeOffset != 0 {
         hasCTO = true
         break
      }
   }
   if !hasCTO {
      return nil // no ctts box needed if all offsets are 0
   }

   var entries []CttsEntry
   if len(samples) > 0 {
      currentOffset := samples[0].CompositionTimeOffset
      currentCount := uint32(0)
      for _, sample := range samples {
         if sample.CompositionTimeOffset == currentOffset {
            currentCount++
         } else {
            entries = append(entries, CttsEntry{currentCount, currentOffset})
            currentOffset = sample.CompositionTimeOffset
            currentCount = 1
         }
      }
      entries = append(entries, CttsEntry{currentCount, currentOffset})
   }
   return CttsBox{Entries: entries}.Encode()
}

func buildStsc(counts []uint32) []byte {
   var entries []StscEntry
   chunkIdx := uint32(1)
   for _, count := range counts {
      if n := len(entries); n > 0 && entries[n-1].SamplesPerChunk == count {
         chunkIdx++
         continue
      }
      entries = append(entries, StscEntry{chunkIdx, count, 1})
      chunkIdx++
   }
   return StscBox{Entries: entries}.Encode()
}

func buildStss(samples []*RemuxSample) []byte {
   var indices []uint32
   for i, sample := range samples {
      if sample.IsSync {
         indices = append(indices, uint32(i+1))
      }
   }
   if len(indices) == len(samples) {
      return nil
   }
   return StssBox{Indices: indices}.Encode()
}

func buildStsz(samples []*RemuxSample) []byte {
   entries := make([]uint32, len(samples))
   for i, sample := range samples {
      entries[i] = sample.Size
   }
   return StszBox{SampleCount: uint32(len(samples)), EntrySizes: entries}.Encode()
}

func buildStts(samples []*RemuxSample) []byte {
   if len(samples) == 0 {
      return nil
   }
   var entries []SttsEntry
   currentDuration := samples[0].Duration
   currentCount := uint32(0)
   for _, sample := range samples {
      if sample.Duration == currentDuration {
         currentCount++
      } else {
         entries = append(entries, SttsEntry{currentCount, currentDuration})
         currentDuration = sample.Duration
         currentCount = 1
      }
   }
   entries = append(entries, SttsEntry{currentCount, currentDuration})
   return SttsBox{Entries: entries}.Encode()
}

// This file is the two directions between remuxer state and the sample
// tables of a moov: StateFromMoov recovers state from the tables a
// previous Finish wrote, and the build functions turn state back into the
// table boxes Finish appends to the file.

// RemuxState is the remuxer bookkeeping recovered from a stopped file. It
// is exactly the state the remuxer held when it was stopped, derived from
// the sample tables in the moov that Finish wrote to the file.
type RemuxState struct {
   Samples         []*RemuxSample
   ChunkOffsets    []uint64
   SamplesPerChunk []uint32
}

// StateFromMoov rebuilds remuxer state from the moov that Finish wrote to
// a stopped file. It is a pure decode: the input is only the moov bytes,
// which the caller reads from the file without touching the media data,
// so memory use is bounded by the moov size, never the file size.
func StateFromMoov(moovData []byte) (*RemuxState, error) {
   moov, err := DecodeMoovBox(moovData)
   if err != nil {
      return nil, fmt.Errorf("parsing moov: %w", err)
   }
   if len(moov.Trak) == 0 {
      return nil, errors.New("no trak in moov")
   }
   trak := moov.Trak[0]
   if trak.Mdia == nil || trak.Mdia.Minf == nil || trak.Mdia.Minf.Stbl == nil {
      return nil, errors.New("malformed trak in moov")
   }
   stbl := trak.Mdia.Minf.Stbl
   if stbl.Stsz == nil || stbl.Stsc == nil {
      return nil, errors.New("moov is missing sample tables")
   }
   if stbl.Stco == nil && stbl.Co64 == nil {
      return nil, errors.New("moov is missing chunk offsets")
   }

   var chunkOffsets []uint64
   if stbl.Stco != nil {
      chunkOffsets = make([]uint64, len(stbl.Stco.Offsets))
      for i, offset := range stbl.Stco.Offsets {
         chunkOffsets[i] = uint64(offset)
      }
   } else {
      chunkOffsets = append([]uint64{}, stbl.Co64.Offsets...)
   }

   // Expand the run-length stsc into one sample count per chunk.
   if len(stbl.Stsc.Entries) == 0 && len(chunkOffsets) > 0 {
      return nil, errors.New("stsc is empty but chunks exist")
   }
   samplesPerChunk := make([]uint32, len(chunkOffsets))
   entryIndex := 0
   for i := range chunkOffsets {
      chunkNumber := uint32(i + 1)
      for entryIndex+1 < len(stbl.Stsc.Entries) &&
         stbl.Stsc.Entries[entryIndex+1].FirstChunk <= chunkNumber {
         entryIndex++
      }
      samplesPerChunk[i] = stbl.Stsc.Entries[entryIndex].SamplesPerChunk
   }

   // Sample count and sizes. A non-zero SampleSize means a constant size
   // with no per-sample entries.
   count := len(stbl.Stsz.EntrySizes)
   if stbl.Stsz.SampleSize != 0 {
      count = int(stbl.Stsz.SampleCount)
   }
   samples := make([]*RemuxSample, count)
   for i := range samples {
      samples[i] = &RemuxSample{IsSync: true}
      if stbl.Stsz.SampleSize != 0 {
         samples[i].Size = stbl.Stsz.SampleSize
      } else {
         samples[i].Size = stbl.Stsz.EntrySizes[i]
      }
   }

   // Durations, run-length expanded.
   if stbl.Stts != nil {
      var total uint32
      for _, entry := range stbl.Stts.Entries {
         total += entry.SampleCount
      }
      if int(total) != count {
         return nil, errors.New("stts does not cover every sample")
      }
      index := 0
      for _, entry := range stbl.Stts.Entries {
         for j := uint32(0); j < entry.SampleCount; j++ {
            samples[index].Duration = entry.SampleDuration
            index++
         }
      }
   } else if count > 0 {
      return nil, errors.New("moov is missing sample durations")
   }

   // Sync samples. An absent stss means every sample is a sync sample.
   if stbl.Stss != nil {
      for _, sample := range samples {
         sample.IsSync = false
      }
      for _, index := range stbl.Stss.Indices {
         if index == 0 || int(index) > count {
            return nil, errors.New("stss index out of range")
         }
         samples[index-1].IsSync = true
      }
   }

   // Composition time offsets. An absent ctts means every offset is zero.
   if stbl.Ctts != nil {
      var total uint32
      for _, entry := range stbl.Ctts.Entries {
         total += entry.SampleCount
      }
      if int(total) != count {
         return nil, errors.New("ctts does not cover every sample")
      }
      index := 0
      for _, entry := range stbl.Ctts.Entries {
         for j := uint32(0); j < entry.SampleCount; j++ {
            samples[index].CompositionTimeOffset = entry.SampleOffset
            index++
         }
      }
   }

   // Sanity: the per-chunk sample counts must add up to the sample list.
   var totalChunks uint32
   for _, perChunk := range samplesPerChunk {
      totalChunks += perChunk
   }
   if int(totalChunks) != count {
      return nil, errors.New("stsc does not cover every sample")
   }

   return &RemuxState{
      Samples:         samples,
      ChunkOffsets:    chunkOffsets,
      SamplesPerChunk: samplesPerChunk,
   }, nil
}

// state.go
