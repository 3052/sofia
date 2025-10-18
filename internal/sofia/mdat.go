// File: mdat_box.go
package mp4parser

// MdatBox (Media Data Box)
type MdatBox struct {
	Data []byte
}

func ParseMdatBox(data []byte) (*MdatBox, error) {
	return &MdatBox{Data: data}, nil
}

func (b *MdatBox) Size() uint64 {
	size := uint64(8 + len(b.Data))
	if size > 0xFFFFFFFF {
		return size + 8 
	}
	return size
}

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