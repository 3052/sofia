package sofia

import (
   "encoding/binary"
)

// --- Box Builders ---

func buildStts(samples []sampleInfo) []byte {
   if len(samples) == 0 {
      return nil
   }
   var entries []byte
   count := uint32(0)
   currDur := samples[0].Duration
   currCnt := uint32(0)

   for _, s := range samples {
      if s.Duration == currDur {
         currCnt++
      } else {
         entries = append(entries, uint32ToBytes(currCnt)...)
         entries = append(entries, uint32ToBytes(currDur)...)
         count++
         currDur = s.Duration
         currCnt = 1
      }
   }
   // Flush last entry
   entries = append(entries, uint32ToBytes(currCnt)...)
   entries = append(entries, uint32ToBytes(currDur)...)
   count++

   data := make([]byte, 8) // Version(0) + Flags(0) + EntryCount
   binary.BigEndian.PutUint32(data[4:8], count)
   return makeBox("stts", append(data, entries...))
}

func buildStsz(samples []sampleInfo) []byte {
   // Version(0) + Flags(0) + SampleSize(0=var) + SampleCount
   data := make([]byte, 12)
   binary.BigEndian.PutUint32(data[8:12], uint32(len(samples)))

   entries := make([]byte, len(samples)*4)
   for i, s := range samples {
      binary.BigEndian.PutUint32(entries[i*4:], s.Size)
   }
   return makeBox("stsz", append(data, entries...))
}

func buildStsc(counts []uint32) []byte {
   // Maps: 1 Chunk = 1 Segment.
   // We group consecutive chunks that have the same number of samples.
   type stscEntry struct {
      firstChunk      uint32
      samplesPerChunk uint32
      sdi             uint32
   }
   var list []stscEntry

   chunkIdx := uint32(1) // 1-based index
   for _, cnt := range counts {
      if len(list) > 0 {
         last := &list[len(list)-1]
         if last.samplesPerChunk == cnt {
            // This chunk looks just like the previous run
            chunkIdx++
            continue
         }
      }
      list = append(list, stscEntry{firstChunk: chunkIdx, samplesPerChunk: cnt, sdi: 1})
      chunkIdx++
   }

   data := make([]byte, 8)
   binary.BigEndian.PutUint32(data[4:8], uint32(len(list)))
   entries := make([]byte, len(list)*12)
   for i, e := range list {
      off := i * 12
      binary.BigEndian.PutUint32(entries[off:], e.firstChunk)
      binary.BigEndian.PutUint32(entries[off+4:], e.samplesPerChunk)
      binary.BigEndian.PutUint32(entries[off+8:], e.sdi)
   }
   return makeBox("stsc", append(data, entries...))
}

func buildStss(samples []sampleInfo) []byte {
   var indices []uint32
   allSync := true
   for i, s := range samples {
      if s.IsSync {
         indices = append(indices, uint32(i+1)) // 1-based index
      } else {
         allSync = false
      }
   }
   // If every single sample is a sync sample, stss is not required.
   if allSync {
      return nil
   }

   data := make([]byte, 8)
   binary.BigEndian.PutUint32(data[4:8], uint32(len(indices)))
   entries := make([]byte, len(indices)*4)
   for i, idx := range indices {
      binary.BigEndian.PutUint32(entries[i*4:], idx)
   }
   return makeBox("stss", append(data, entries...))
}

func buildCtts(samples []sampleInfo) []byte {
   hasCtts := false
   for _, s := range samples {
      if s.CompositionOffset != 0 {
         hasCtts = true
         break
      }
   }
   if !hasCtts {
      return nil
   }

   // RLE compression for offsets
   var entries []byte
   count := uint32(0)
   currOff := samples[0].CompositionOffset
   currCnt := uint32(0)

   for _, s := range samples {
      if s.CompositionOffset == currOff {
         currCnt++
      } else {
         entries = append(entries, uint32ToBytes(currCnt)...)
         // Cast int32 to uint32 for binary writing
         entries = append(entries, uint32ToBytes(uint32(currOff))...)
         count++
         currOff = s.CompositionOffset
         currCnt = 1
      }
   }
   // Flush last
   entries = append(entries, uint32ToBytes(currCnt)...)
   entries = append(entries, uint32ToBytes(uint32(currOff))...)
   count++

   data := make([]byte, 8)
   binary.BigEndian.PutUint32(data[4:8], count)
   return makeBox("ctts", append(data, entries...))
}

func buildStco(offsets []uint64) []byte {
   data := make([]byte, 8)
   binary.BigEndian.PutUint32(data[4:8], uint32(len(offsets)))
   entries := make([]byte, len(offsets)*4)
   for i, o := range offsets {
      binary.BigEndian.PutUint32(entries[i*4:], uint32(o))
   }
   return makeBox("stco", append(data, entries...))
}

func buildCo64(offsets []uint64) []byte {
   data := make([]byte, 8)
   binary.BigEndian.PutUint32(data[4:8], uint32(len(offsets)))
   entries := make([]byte, len(offsets)*8)
   for i, o := range offsets {
      binary.BigEndian.PutUint64(entries[i*8:], o)
   }
   return makeBox("co64", append(data, entries...))
}

// --- Utility Helpers ---

func makeBox(typeStr string, payload []byte) []byte {
   size := 8 + len(payload)
   buf := make([]byte, 8)
   binary.BigEndian.PutUint32(buf[0:4], uint32(size))
   copy(buf[4:8], []byte(typeStr))
   return append(buf, payload...)
}

func uint32ToBytes(v uint32) []byte {
   b := make([]byte, 4)
   binary.BigEndian.PutUint32(b, v)
   return b
}

// FindMoofPtr finds the first MoofBox pointer in a slice of generic boxes.
func FindMoofPtr(boxes []Box) *MoofBox {
   for _, box := range boxes {
      if box.Moof != nil {
         return box.Moof
      }
   }
   return nil
}

// FindMdatPtr finds the first MdatBox pointer in a slice of generic boxes.
func FindMdatPtr(boxes []Box) *MdatBox {
   for _, box := range boxes {
      if box.Mdat != nil {
         return box.Mdat
      }
   }
   return nil
}

// filterMvex removes the 'mvex' atom from the MoovBox children.
// The mvex atom signals fragmentation; removing it signals a regular MP4.
func filterMvex(moov *MoovBox) {
   var cleanChildren []MoovChild
   for _, child := range moov.Children {
      // In sofia, unknown boxes are often stored in 'Raw'.
      // mvex is usually Type "mvex" (0x6d766578).
      // Fix: Removed 'child.Raw != nil' check (S1009)
      if len(child.Raw) >= 8 {
         if string(child.Raw[4:8]) == "mvex" {
            continue
         }
      }
      cleanChildren = append(cleanChildren, child)
   }
   moov.Children = cleanChildren
}
