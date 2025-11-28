package sofia

import "encoding/binary"

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
