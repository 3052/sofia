// File: free_box.go
package mp4parser

// FreeBox (Free Space Box)
type FreeBox struct {
	Data []byte
}

// ParseFreeBox parses the FreeBox from its content slice.
func ParseFreeBox(data []byte) (*FreeBox, error) {
	return &FreeBox{Data: data}, nil
}

// Size calculates the total byte size of the FreeBox.
func (b *FreeBox) Size() uint64 {
	return uint64(8 + len(b.Data))
}

// Format serializes the FreeBox into the destination slice and returns the new offset.
func (b *FreeBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "free")
	offset = writeBytes(dst, offset, b.Data)
	return offset
}
