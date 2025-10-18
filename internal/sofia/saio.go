// File: saio_box.go
package mp4parser

// SaioBox (Sample Auxiliary Information Offsets Box)
type SaioBox struct {
	FullBox
	Offsets []uint64
}

// ParseSaioBox parses the SaioBox from its content slice.
func ParseSaioBox(data []byte) (*SaioBox, error) {
	b := &SaioBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	var entryCount uint32
	entryCount, offset, err = readUint32(data, offset)
	if err != nil {
		return nil, err
	}
	b.Offsets = make([]uint64, entryCount)
	for i := 0; i < int(entryCount); i++ {
		if b.Version == 1 {
			var val uint64
			val, offset, err = readUint64(data, offset)
			if err != nil {
				return nil, err
			}
			b.Offsets[i] = val
		} else {
			var val uint32
			val, offset, err = readUint32(data, offset)
			if err != nil {
				return nil, err
			}
			b.Offsets[i] = uint64(val)
		}
	}
	return b, nil
}

// Size calculates the total byte size of the SaioBox.
func (b *SaioBox) Size() uint64 {
	size := uint64(8) // Header
	size += b.FullBox.Size()
	size += 4 // EntryCount
	if b.Version == 1 {
		size += uint64(len(b.Offsets) * 8)
	} else {
		size += uint64(len(b.Offsets) * 4)
	}
	return size
}

// Format serializes the SaioBox into the destination slice and returns the new offset.
func (b *SaioBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "saio")
	offset = b.FullBox.Format(dst, offset)
	offset = writeUint32(dst, offset, uint32(len(b.Offsets)))
	for _, o := range b.Offsets {
		if b.Version == 1 {
			offset = writeUint64(dst, offset, o)
		} else {
			offset = writeUint32(dst, offset, uint32(o))
		}
	}
	return offset
}
