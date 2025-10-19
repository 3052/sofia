// File: full_box.go
package mp4parser

type FullBox struct {
   Version uint8
   Flags   [3]byte
}

func (b *FullBox) Parse(data []byte, offset int) (int, error) {
   if offset+4 > len(data) {
      return offset, ErrUnexpectedEOF
   }
   b.Version = data[offset]
   copy(b.Flags[:], data[offset+1:offset+4])
   return offset + 4, nil
}
func (b *FullBox) Size() uint64 { return 4 }
func (b *FullBox) Format(dst []byte, offset int) int {
   offset = writeUint8(dst, offset, b.Version)
   offset = writeBytes(dst, offset, b.Flags[:])
   return offset
}
