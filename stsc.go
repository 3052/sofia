package sofia

import "encoding/binary"

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
            // Continuation of previous run
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
