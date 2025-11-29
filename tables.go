package sofia

import "encoding/binary"

// -----------------------------------------------------------------------------
// STTS (Decoding Time to Sample)
// -----------------------------------------------------------------------------

type SttsEntry struct {
   SampleCount    uint32
   SampleDuration uint32
}

type SttsBox struct {
   Header  BoxHeader
   Entries []SttsEntry
}

func (b *SttsBox) Encode() []byte {
   // 8 bytes (Ver/Flags + Count) + (Entries * 8 bytes)
   entryCount := uint32(len(b.Entries))
   contentSize := 8 + (entryCount * 8)
   b.Header.Size = 8 + contentSize
   b.Header.Type = [4]byte{'s', 't', 't', 's'}

   buf := make([]byte, b.Header.Size)
   copy(buf[0:8], b.Header.Encode())

   // Version(1) + Flags(3) + EntryCount(4)
   binary.BigEndian.PutUint32(buf[12:16], entryCount)

   offset := 16
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(buf[offset:offset+4], entry.SampleCount)
      binary.BigEndian.PutUint32(buf[offset+4:offset+8], entry.SampleDuration)
      offset += 8
   }
   return buf
}

// buildStts constructs the box by RLE-compressing the sample durations.
func buildStts(samples []sampleInfo) []byte {
   if len(samples) == 0 {
      return nil
   }
   var entries []SttsEntry
   currDur := samples[0].Duration
   currCnt := uint32(0)

   for _, s := range samples {
      if s.Duration == currDur {
         currCnt++
      } else {
         entries = append(entries, SttsEntry{SampleCount: currCnt, SampleDuration: currDur})
         currDur = s.Duration
         currCnt = 1
      }
   }
   // Flush last
   entries = append(entries, SttsEntry{SampleCount: currCnt, SampleDuration: currDur})

   box := SttsBox{Entries: entries}
   return box.Encode()
}

// -----------------------------------------------------------------------------
// STSZ (Sample Sizes)
// -----------------------------------------------------------------------------

type StszBox struct {
   Header      BoxHeader
   SampleSize  uint32 // 0 means variable sizes follow
   SampleCount uint32
   EntrySizes  []uint32
}

func (b *StszBox) Encode() []byte {
   // Header(8) + Ver/Flags(4) + SampleSize(4) + Count(4) + (Entries * 4)
   contentSize := 12 + (uint32(len(b.EntrySizes)) * 4)
   b.Header.Size = 8 + contentSize
   b.Header.Type = [4]byte{'s', 't', 's', 'z'}

   buf := make([]byte, b.Header.Size)
   copy(buf[0:8], b.Header.Encode())

   // Offset 8: Version/Flags (0)
   binary.BigEndian.PutUint32(buf[12:16], b.SampleSize)
   binary.BigEndian.PutUint32(buf[16:20], b.SampleCount)

   offset := 20
   for _, size := range b.EntrySizes {
      binary.BigEndian.PutUint32(buf[offset:offset+4], size)
      offset += 4
   }
   return buf
}

func buildStsz(samples []sampleInfo) []byte {
   entries := make([]uint32, len(samples))
   for i, s := range samples {
      entries[i] = s.Size
   }
   box := StszBox{
      SampleSize:  0, // Variable sizes
      SampleCount: uint32(len(samples)),
      EntrySizes:  entries,
   }
   return box.Encode()
}

// -----------------------------------------------------------------------------
// STSC (Sample To Chunk)
// -----------------------------------------------------------------------------

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
   // Header(8) + Ver/Flags(4) + Count(4) + (Entries * 12)
   entryCount := uint32(len(b.Entries))
   contentSize := 8 + (entryCount * 12)
   b.Header.Size = 8 + contentSize
   b.Header.Type = [4]byte{'s', 't', 's', 'c'}

   buf := make([]byte, b.Header.Size)
   copy(buf[0:8], b.Header.Encode())

   // Offset 8: Version/Flags (0)
   binary.BigEndian.PutUint32(buf[12:16], entryCount)

   offset := 16
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(buf[offset:offset+4], entry.FirstChunk)
      binary.BigEndian.PutUint32(buf[offset+4:offset+8], entry.SamplesPerChunk)
      binary.BigEndian.PutUint32(buf[offset+8:offset+12], entry.SampleDescriptionIndex)
      offset += 12
   }
   return buf
}

func buildStsc(counts []uint32) []byte {
   var entries []StscEntry
   chunkIdx := uint32(1) // 1-based index

   for _, cnt := range counts {
      if len(entries) > 0 {
         last := &entries[len(entries)-1]
         if last.SamplesPerChunk == cnt {
            // Continuation of previous run.
            // We MUST still increment the chunk index counter!
            chunkIdx++
            continue
         }
      }
      entries = append(entries, StscEntry{
         FirstChunk:             chunkIdx,
         SamplesPerChunk:        cnt,
         SampleDescriptionIndex: 1,
      })
      chunkIdx++
   }

   box := StscBox{Entries: entries}
   return box.Encode()
}

// -----------------------------------------------------------------------------
// STCO (Chunk Offset - 32-bit)
// -----------------------------------------------------------------------------

type StcoBox struct {
   Header  BoxHeader
   Offsets []uint32
}

func (b *StcoBox) Encode() []byte {
   // Header(8) + Ver/Flags(4) + Count(4) + (Entries * 4)
   entryCount := uint32(len(b.Offsets))
   contentSize := 8 + (entryCount * 4)
   b.Header.Size = 8 + contentSize
   b.Header.Type = [4]byte{'s', 't', 'c', 'o'}

   buf := make([]byte, b.Header.Size)
   copy(buf[0:8], b.Header.Encode())

   // Offset 8: Version/Flags (0)
   binary.BigEndian.PutUint32(buf[12:16], entryCount)

   offset := 16
   for _, o := range b.Offsets {
      binary.BigEndian.PutUint32(buf[offset:offset+4], o)
      offset += 4
   }
   return buf
}

func buildStco(offsets []uint64) []byte {
   entries := make([]uint32, len(offsets))
   for i, o := range offsets {
      // Truncate to 32-bit (unsafe if > 4GB)
      entries[i] = uint32(o)
   }
   box := StcoBox{Offsets: entries}
   return box.Encode()
}
