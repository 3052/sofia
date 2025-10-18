package mp4parser

// MfhdBox (Movie Fragment Header Box)
type MfhdBox struct {
	FullBox
	SequenceNumber uint32
}

// ParseMfhdBox parses the MfhdBox from its content slice.
func ParseMfhdBox(data []byte) (*MfhdBox, error) {
	b := &MfhdBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	b.SequenceNumber, _, err = readUint32(data, offset)
	return b, err
}

// Size calculates the total byte size of the MfhdBox.
func (b *MfhdBox) Size() uint64 {
	// 8 (header) + 4 (fullbox) + 4 (seq number)
	return 16
}

// Format serializes the MfhdBox into the destination slice and returns the new offset.
func (b *MfhdBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "mfhd")
	offset = b.FullBox.Format(dst, offset)
	offset = writeUint32(dst, offset, b.SequenceNumber)
	return offset
}
