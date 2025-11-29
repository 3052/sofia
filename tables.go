package sofia

import "encoding/binary"

// -----------------------------------------------------------------------------
// STTS
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
   buf := make([]byte, 16) // Header(8) + Ver/Flags(4) + Count(4)

   // Version(0) + Flags(0)
   // EntryCount
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Entries)))

   tmp := make([]byte, 8)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tmp[0:4], entry.SampleCount)
      binary.BigEndian.PutUint32(tmp[4:8], entry.SampleDuration)
      buf = append(buf, tmp...)
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 't', 's'}
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])

   return buf
}

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
         entries = append(entries, SttsEntry{currCnt, currDur})
         currDur = s.Duration
         currCnt = 1
      }
   }
   entries = append(entries, SttsEntry{currCnt, currDur})

   box := SttsBox{Entries: entries}
   return box.Encode()
}

// -----------------------------------------------------------------------------
// STSZ
// -----------------------------------------------------------------------------

type StszBox struct {
   Header      BoxHeader
   SampleSize  uint32
   SampleCount uint32
   EntrySizes  []uint32
}

func (b *StszBox) Encode() []byte {
   buf := make([]byte, 20) // Header(8) + Ver/Flags(4) + Size(4) + Count(4)

   binary.BigEndian.PutUint32(buf[12:16], b.SampleSize)
   binary.BigEndian.PutUint32(buf[16:20], b.SampleCount)

   tmp := make([]byte, 4)
   for _, size := range b.EntrySizes {
      binary.BigEndian.PutUint32(tmp, size)
      buf = append(buf, tmp...)
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 's', 'z'}
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])

   return buf
}

func buildStsz(samples []sampleInfo) []byte {
   entries := make([]uint32, len(samples))
   for i, s := range samples {
      entries[i] = s.Size
   }
   box := StszBox{
      SampleSize:  0,
      SampleCount: uint32(len(samples)),
      EntrySizes:  entries,
   }
   return box.Encode()
}

// -----------------------------------------------------------------------------
// STSC
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
   buf := make([]byte, 16) // Header(8) + Ver/Flags(4) + Count(4)

   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Entries)))

   tmp := make([]byte, 12)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tmp[0:4], entry.FirstChunk)
      binary.BigEndian.PutUint32(tmp[4:8], entry.SamplesPerChunk)
      binary.BigEndian.PutUint32(tmp[8:12], entry.SampleDescriptionIndex)
      buf = append(buf, tmp...)
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 's', 'c'}
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])

   return buf
}

func buildStsc(counts []uint32) []byte {
   var entries []StscEntry
   chunkIdx := uint32(1)

   for _, cnt := range counts {
      if len(entries) > 0 {
         last := &entries[len(entries)-1]
         if last.SamplesPerChunk == cnt {
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
// STCO
// -----------------------------------------------------------------------------

type StcoBox struct {
   Header  BoxHeader
   Offsets []uint32
}

func (b *StcoBox) Encode() []byte {
   buf := make([]byte, 16)
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Offsets)))

   tmp := make([]byte, 4)
   for _, o := range b.Offsets {
      binary.BigEndian.PutUint32(tmp, o)
      buf = append(buf, tmp...)
   }

   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 'c', 'o'}
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])

   return buf
}

func buildStco(offsets []uint64) []byte {
   entries := make([]uint32, len(offsets))
   for i, o := range offsets {
      entries[i] = uint32(o)
   }
   box := StcoBox{Offsets: entries}
   return box.Encode()
}
