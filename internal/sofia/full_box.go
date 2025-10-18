package mp4parser

// FullBox is a base struct for boxes that contain a version and flags.
type FullBox struct {
   Version uint8
   Flags   [3]byte
}

// Parse reads the version and flags.
func (b *FullBox) Parse(data []byte, offset int) (int, error) {
   if offset+4 > len(data) {
      return offset, ErrUnexpectedEOF
   }
   b.Version = data[offset]
   copy(b.Flags[:], data[offset+1:offset+4])
   return offset + 4, nil
}

// Size returns the byte size of the full box data.
func (b *FullBox) Size() uint64 {
   return 4
}

// Format writes the FullBox data into the destination slice and returns the new offset.
func (b *FullBox) Format(dst []byte, offset int) int {
   offset = writeUint8(dst, offset, b.Version)
   offset = writeBytes(dst, offset, b.Flags[:])
   return offset
}
