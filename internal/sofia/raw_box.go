// File: raw_box.go
package mp4parser

// RawBox is a generic container for any box whose internal structure we don't need to parse.
type RawBox struct {
	Type    string
	Content []byte
}

// ParseRawBox creates a RawBox from a type and its content slice.
func ParseRawBox(boxType string, data []byte) (*RawBox, error) {
	return &RawBox{Type: boxType, Content: data}, nil
}

// Size calculates the total byte size of the RawBox.
func (b *RawBox) Size() uint64 {
	return uint64(8 + len(b.Content))
}

// Format serializes the RawBox into the destination slice.
func (b *RawBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, b.Type)
	offset = writeBytes(dst, offset, b.Content)
	return offset
}