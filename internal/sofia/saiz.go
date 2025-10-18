// File: saiz_box.go
package mp4parser

// SaizBox (Sample Auxiliary Information Sizes Box)
type SaizBox struct {
	FullBox
	DefaultSampleInfoSize uint8
	SampleCount           uint32
	SampleInfoSize        []uint8 // optional
}

// ParseSaizBox parses the SaizBox from its content slice.
func ParseSaizBox(data []byte) (*SaizBox, error) {
	b := &SaizBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	b.DefaultSampleInfoSize, offset, err = readUint8(data, offset)
	if err != nil {
		return nil, err
	}
	b.SampleCount, offset, err = readUint32(data, offset)
	if err != nil {
		return nil, err
	}
	if b.DefaultSampleInfoSize == 0 {
		count := int(b.SampleCount)
		if offset+count > len(data) {
			return nil, ErrUnexpectedEOF
		}
		b.SampleInfoSize = make([]uint8, count)
		copy(b.SampleInfoSize, data[offset:offset+count])
	}
	return b, nil
}

// Size calculates the total byte size of the SaizBox.
func (b *SaizBox) Size() uint64 {
	size := uint64(8) // Header
	size += b.FullBox.Size()
	size += 1 // DefaultSampleInfoSize
	size += 4 // SampleCount
	if b.DefaultSampleInfoSize == 0 {
		size += uint64(len(b.SampleInfoSize))
	}
	return size
}

// Format serializes the SaizBox into the destination slice and returns the new offset.
func (b *SaizBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "saiz")
	offset = b.FullBox.Format(dst, offset)
	offset = writeUint8(dst, offset, b.DefaultSampleInfoSize)
	offset = writeUint32(dst, offset, b.SampleCount)
	if b.DefaultSampleInfoSize == 0 {
		offset = writeBytes(dst, offset, b.SampleInfoSize)
	}
	return offset
}
