// File: raw_box.go
package mp4parser

type RawBox struct {
   Type    string
   Content []byte
}

func ParseRawBox(boxType string, data []byte) (*RawBox, error) {
   return &RawBox{Type: boxType, Content: data}, nil
}
func (b *RawBox) Size() uint64 { return uint64(8 + len(b.Content)) }
func (b *RawBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, b.Type)
   offset = writeBytes(dst, offset, b.Content)
   return offset
}
