package mp4parser

// SencBox (Sample Encryption Box)
type SencBox struct {
	FullBox
	SampleCount           uint32
	InitializationVectors []InitializationVector
}

// InitializationVector represents the IV and subsample encryption data.
type InitializationVector struct {
	IV         []byte
	Subsamples []Subsample
}

// Subsample defines clear and encrypted byte counts.
type Subsample struct {
	BytesOfClearData     uint16
	BytesOfProtectedData uint32
}

// ParseSencBox parses the SencBox from its content slice.
func ParseSencBox(data []byte) (*SencBox, error) {
	b := &SencBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	b.SampleCount, offset, err = readUint32(data, offset)
	if err != nil {
		return nil, err
	}
	flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
	hasSubsamples := (flags & 0x000002) != 0
	b.InitializationVectors = make([]InitializationVector, b.SampleCount)
	for i := 0; i < int(b.SampleCount); i++ {
		iv := InitializationVector{}
		ivSize := 8
		if offset+ivSize > len(data) {
			return nil, ErrUnexpectedEOF
		}
		iv.IV = data[offset : offset+ivSize]
		offset += ivSize
		if hasSubsamples {
			var subsampleCount uint16
			subsampleCount, offset, err = readUint16(data, offset)
			if err != nil {
				return nil, err
			}
			iv.Subsamples = make([]Subsample, subsampleCount)
			for j := 0; j < int(subsampleCount); j++ {
				var clearData uint16
				clearData, offset, err = readUint16(data, offset)
				if err != nil {
					return nil, err
				}
				var protectedData uint32
				protectedData, offset, err = readUint32(data, offset)
				if err != nil {
					return nil, err
				}
				iv.Subsamples[j] = Subsample{
					BytesOfClearData:     clearData,
					BytesOfProtectedData: protectedData,
				}
			}
		}
		b.InitializationVectors[i] = iv
	}
	return b, nil
}

// Size calculates the total byte size of the SencBox.
func (b *SencBox) Size() uint64 {
	size := uint64(8) // Header
	size += b.FullBox.Size()
	size += 4 // SampleCount
	flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
	hasSubsamples := (flags & 0x000002) != 0
	for _, iv := range b.InitializationVectors {
		size += uint64(len(iv.IV))
		if hasSubsamples {
			size += 2 // Subsample count
			size += uint64(len(iv.Subsamples) * (2 + 4)) // clear + protected
		}
	}
	return size
}

// Format serializes the SencBox into the destination slice and returns the new offset.
func (b *SencBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "senc")
	offset = b.FullBox.Format(dst, offset)
	offset = writeUint32(dst, offset, b.SampleCount)
	flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
	hasSubsamples := (flags & 0x000002) != 0
	for _, iv := range b.InitializationVectors {
		offset = writeBytes(dst, offset, iv.IV)
		if hasSubsamples {
			offset = writeUint16(dst, offset, uint16(len(iv.Subsamples)))
			for _, sub := range iv.Subsamples {
				offset = writeUint16(dst, offset, sub.BytesOfClearData)
				offset = writeUint32(dst, offset, sub.BytesOfProtectedData)
			}
		}
	}
	return offset
}
