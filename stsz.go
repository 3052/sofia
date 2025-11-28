package sofia

import "encoding/binary"

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
