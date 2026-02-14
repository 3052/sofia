package sofia

import (
   "errors"
)

// --- STBL ---
type StblBox struct {
   Header      BoxHeader
   Stsd        *StsdBox
   RawChildren [][]byte
}

func (b *StblBox) Parse(data []byte) error {
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
      case "stsd":
         var stsd StsdBox
         if err := stsd.Parse(content); err != nil {
            return err
         }
         b.Stsd = &stsd
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return nil
}

func (b *StblBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Stsd != nil {
      buffer = append(buffer, b.Stsd.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- STTS ---
type SttsEntry struct {
   SampleCount    uint32
   SampleDuration uint32
}

type SttsBox struct {
   Header  BoxHeader
   Entries []SttsEntry
}

func (b *SttsBox) Encode() []byte {
   size := 16 + len(b.Entries)*8
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip header
   w.PutBytes([]byte{0, 0, 0, 0})
   w.PutUint32(uint32(len(b.Entries)))
   for _, entry := range b.Entries {
      w.PutUint32(entry.SampleCount)
      w.PutUint32(entry.SampleDuration)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'s', 't', 't', 's'}
   b.Header.Put(buffer)
   return buffer
}

func buildStts(samples []RemuxSample) []byte {
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
   box := SttsBox{Entries: entries}
   return box.Encode()
}

// --- CTTS ---
type CttsEntry struct {
   SampleCount  uint32
   SampleOffset int32
}

type CttsBox struct {
   Header  BoxHeader
   Entries []CttsEntry
}

func (b *CttsBox) Encode() []byte {
   size := 16 + len(b.Entries)*8
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8   // Skip header
   w.PutUint32(0) // Version 0, Flags 0
   w.PutUint32(uint32(len(b.Entries)))
   for _, entry := range b.Entries {
      w.PutUint32(entry.SampleCount)
      w.PutUint32(uint32(entry.SampleOffset))
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'c', 't', 't', 's'}
   b.Header.Put(buffer)
   return buffer
}

func buildCtts(samples []RemuxSample) []byte {
   hasCTO := false
   for _, sample := range samples {
      if sample.CompositionTimeOffset != 0 {
         hasCTO = true
         break
      }
   }
   if !hasCTO {
      return nil // No ctts box needed if all offsets are 0
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

   box := CttsBox{Entries: entries}
   return box.Encode()
}

// --- STSZ ---
type StszBox struct {
   Header      BoxHeader
   SampleSize  uint32
   SampleCount uint32
   EntrySizes  []uint32
}

func (b *StszBox) Encode() []byte {
   size := 20 + len(b.EntrySizes)*4
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip header
   w.PutUint32(0)
   w.PutUint32(b.SampleSize)
   w.PutUint32(b.SampleCount)
   for _, entrySize := range b.EntrySizes {
      w.PutUint32(entrySize)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'s', 't', 's', 'z'}
   b.Header.Put(buffer)
   return buffer
}

func buildStsz(samples []RemuxSample) []byte {
   entries := make([]uint32, len(samples))
   for i, sample := range samples {
      entries[i] = sample.Size
   }
   box := StszBox{SampleSize: 0, SampleCount: uint32(len(samples)), EntrySizes: entries}
   return box.Encode()
}

// --- STSC ---
type StscEntry struct {
   FirstChunk             uint32
   SamplesPerChunk        uint32
   SampleDescriptionIndex uint32
}

type StscBox struct {
   Header  BoxHeader
   Entries []StscEntry
}

func (b *StscBox) Encode() []byte {
   size := 16 + len(b.Entries)*12
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip header
   w.PutUint32(0)
   w.PutUint32(uint32(len(b.Entries)))
   for _, entry := range b.Entries {
      w.PutUint32(entry.FirstChunk)
      w.PutUint32(entry.SamplesPerChunk)
      w.PutUint32(entry.SampleDescriptionIndex)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'s', 't', 's', 'c'}
   b.Header.Put(buffer)
   return buffer
}

func buildStsc(counts []uint32) []byte {
   var entries []StscEntry
   chunkIdx := uint32(1)
   for _, count := range counts {
      if len(entries) > 0 {
         last := &entries[len(entries)-1]
         if last.SamplesPerChunk == count {
            chunkIdx++
            continue
         }
      }
      entries = append(entries, StscEntry{chunkIdx, count, 1})
      chunkIdx++
   }
   box := StscBox{Entries: entries}
   return box.Encode()
}

// --- STCO ---
type StcoBox struct {
   Header  BoxHeader
   Offsets []uint32
}

func (b *StcoBox) Encode() []byte {
   size := 16 + len(b.Offsets)*4
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip header
   w.PutUint32(0)
   w.PutUint32(uint32(len(b.Offsets)))
   for _, offset := range b.Offsets {
      w.PutUint32(offset)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'s', 't', 'c', 'o'}
   b.Header.Put(buffer)
   return buffer
}

// --- CO64 ---
type Co64Box struct {
   Header  BoxHeader
   Offsets []uint64
}

func (b *Co64Box) Encode() []byte {
   size := 16 + len(b.Offsets)*8
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip header
   w.PutUint32(0)
   w.PutUint32(uint32(len(b.Offsets)))
   for _, offset := range b.Offsets {
      w.PutUint64(offset)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'c', 'o', '6', '4'}
   b.Header.Put(buffer)
   return buffer
}

// buildChunkOffsetBox decides whether to use stco or co64.
func buildChunkOffsetBox(offsets []uint64) []byte {
   use64bit := false
   for _, offset := range offsets {
      if offset > 0xFFFFFFFF {
         use64bit = true
         break
      }
   }

   if use64bit {
      // Build a 'co64' box
      box := Co64Box{Offsets: offsets}
      return box.Encode()
   }

   // Build an 'stco' box
   entries32 := make([]uint32, len(offsets))
   for i, offset := range offsets {
      entries32[i] = uint32(offset)
   }
   box := StcoBox{Offsets: entries32}
   return box.Encode()
}

// --- STSS ---
type StssBox struct {
   Header  BoxHeader
   Indices []uint32
}

func (b *StssBox) Encode() []byte {
   size := 16 + len(b.Indices)*4
   buffer := make([]byte, size)
   w := writer{buf: buffer}
   w.offset = 8 // Skip Header
   w.PutUint32(0)
   w.PutUint32(uint32(len(b.Indices)))
   for _, index := range b.Indices {
      w.PutUint32(index)
   }

   b.Header.Size = uint32(size)
   b.Header.Type = [4]byte{'s', 't', 's', 's'}
   b.Header.Put(buffer)
   return buffer
}

func buildStss(samples []RemuxSample) []byte {
   var indices []uint32
   for i, sample := range samples {
      if sample.IsSync {
         indices = append(indices, uint32(i+1))
      }
   }
   if len(indices) == len(samples) {
      return nil
   }
   box := StssBox{Indices: indices}
   return box.Encode()
}
