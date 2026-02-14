package sofia

import "encoding/binary"

// Filename: sofia/tables.go
// Add the following code to this file.

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
   buffer := make([]byte, 16)
   // Version 0, Flags 0
   binary.BigEndian.PutUint32(buffer[8:12], 0)
   binary.BigEndian.PutUint32(buffer[12:16], uint32(len(b.Entries)))
   tempBuffer := make([]byte, 8)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tempBuffer[0:4], entry.SampleCount)
      binary.BigEndian.PutUint32(tempBuffer[4:8], uint32(entry.SampleOffset))
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
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

// ... (rest of tables.go) ...

// --- STBL ---
type StblChild struct {
   Stsd *StsdBox
   Raw  []byte
}

type StblBox struct {
   Header   BoxHeader
   Children []StblChild
}

func (b *StblBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(header BoxHeader, content []byte) error {
      var child StblChild
      switch string(header.Type[:]) {
      case "stsd":
         var stsd StsdBox
         if err := stsd.Parse(content); err != nil {
            return err
         }
         child.Stsd = &stsd
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *StblBox) Encode() []byte {
   buffer := make([]byte, 8)
   for _, child := range b.Children {
      if child.Stsd != nil {
         buffer = append(buffer, child.Stsd.Encode()...)
      } else if child.Raw != nil {
         buffer = append(buffer, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

func (b *StblBox) Stsd() (*StsdBox, bool) {
   for _, child := range b.Children {
      if child.Stsd != nil {
         return child.Stsd, true
      }
   }
   return nil, false
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
   buffer := make([]byte, 16)
   binary.BigEndian.PutUint32(buffer[12:16], uint32(len(b.Entries)))
   tempBuffer := make([]byte, 8)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tempBuffer[0:4], entry.SampleCount)
      binary.BigEndian.PutUint32(tempBuffer[4:8], entry.SampleDuration)
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
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

// --- STSZ ---
type StszBox struct {
   Header      BoxHeader
   SampleSize  uint32
   SampleCount uint32
   EntrySizes  []uint32
}

func (b *StszBox) Encode() []byte {
   buffer := make([]byte, 20)
   binary.BigEndian.PutUint32(buffer[12:16], b.SampleSize)
   binary.BigEndian.PutUint32(buffer[16:20], b.SampleCount)
   tempBuffer := make([]byte, 4)
   for _, size := range b.EntrySizes {
      binary.BigEndian.PutUint32(tempBuffer, size)
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
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
   buffer := make([]byte, 16)
   binary.BigEndian.PutUint32(buffer[12:16], uint32(len(b.Entries)))
   tempBuffer := make([]byte, 12)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tempBuffer[0:4], entry.FirstChunk)
      binary.BigEndian.PutUint32(tempBuffer[4:8], entry.SamplesPerChunk)
      binary.BigEndian.PutUint32(tempBuffer[8:12], entry.SampleDescriptionIndex)
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
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
   buffer := make([]byte, 16)
   binary.BigEndian.PutUint32(buffer[12:16], uint32(len(b.Offsets)))
   tempBuffer := make([]byte, 4)
   for _, offset := range b.Offsets {
      binary.BigEndian.PutUint32(tempBuffer, offset)
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Type = [4]byte{'s', 't', 'c', 'o'}
   b.Header.Put(buffer)
   return buffer
}

func buildStco(offsets []uint64) []byte {
   entries := make([]uint32, len(offsets))
   for i, offset := range offsets {
      entries[i] = uint32(offset)
   }
   box := StcoBox{Offsets: entries}
   return box.Encode()
}

// --- STSS ---
type StssBox struct {
   Header  BoxHeader
   Indices []uint32
}

func (b *StssBox) Encode() []byte {
   buffer := make([]byte, 16)
   binary.BigEndian.PutUint32(buffer[12:16], uint32(len(b.Indices)))
   tempBuffer := make([]byte, 4)
   for _, index := range b.Indices {
      binary.BigEndian.PutUint32(tempBuffer, index)
      buffer = append(buffer, tempBuffer...)
   }
   b.Header.Size = uint32(len(buffer))
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
