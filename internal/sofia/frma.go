// File: frma_box.go
package mp4parser

// FrmaBox (Original Format Box)
type FrmaBox struct {
	DataFormat []byte // 4 bytes
}
func ParseFrmaBox(data []byte) (*FrmaBox, error) {
	b := &FrmaBox{}
	if len(data) < 4 { return nil, ErrUnexpectedEOF }
	b.DataFormat = data[:4]
	return b, nil
}
func (b *FrmaBox) Size() uint64 {
	return 8 + 4
}
func (b *FrmaBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "frma")
	offset = writeBytes(dst, offset, b.DataFormat)
	return offset
}