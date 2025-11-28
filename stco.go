package sofia

import "encoding/binary"

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
