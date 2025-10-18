// File: mdat_box.go
package mp4parser

// MdatBox (Media Data Box)
type MdatBox struct {
	Data []byte
}

// ParseMdatBox parses the MdatBox from its content slice.
func ParseMdatBox(data []byte) (*MdatBox, error) {
	return &MdatBox{Data: data}, nil
}

// Size calculates the total byte size of the MdatBox.
func (b *MdatBox) Size() uint64 {
	size := uint64(8 + len(b.Data))
	if size > 0xFFFFFFFF {
		return size + 8 // Add 8 for largesize field
	}
	return size
}

// Format serializes the MdatBox into the destination slice and returns the new offset.
func (b *MdatBox) Format(dst []byte, offset int) int {
	totalSize := uint64(8 + len(b.Data))
	if totalSize > 0xFFFFFFFF {
		offset = writeUint32(dst, offset, 1)
		offset = writeString(dst, offset, "mdat")
		offset = writeUint64(dst, offset, b.Size())
	} else {
		offset = writeUint32(dst, offset, uint32(b.Size()))
		offset = writeString(dst, offset, "mdat")
	}
	offset = writeBytes(dst, offset, b.Data)
	return offset
}
