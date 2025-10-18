package mp4parser

// TfdtBox (Track Fragment Decode Time Box)
type TfdtBox struct {
	FullBox
	BaseMediaDecodeTime uint64
}

// ParseTfdtBox parses the TfdtBox from its content slice.
func ParseTfdtBox(data []byte) (*TfdtBox, error) {
	b := &TfdtBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	if b.Version == 1 {
		b.BaseMediaDecodeTime, _, err = readUint64(data, offset)
	} else {
		var decodeTime32 uint32
		decodeTime32, _, err = readUint32(data, offset)
		b.BaseMediaDecodeTime = uint64(decodeTime32)
	}
	return b, err
}

// Size calculates the total byte size of the TfdtBox.
func (b *TfdtBox) Size() uint64 {
	size := uint64(8) // Header
	size += b.FullBox.Size()
	if b.Version == 1 {
		size += 8
	} else {
		size += 4
	}
	return size
}

// Format serializes the TfdtBox into the destination slice and returns the new offset.
func (b *TfdtBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "tfdt")
	offset = b.FullBox.Format(dst, offset)
	if b.Version == 1 {
		offset = writeUint64(dst, offset, b.BaseMediaDecodeTime)
	} else {
		offset = writeUint32(dst, offset, uint32(b.BaseMediaDecodeTime))
	}
	return offset
}
