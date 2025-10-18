// File: box_header.go
package mp4parser

// BoxHeader represents the size and type of an MP4 box.
type BoxHeader struct {
	Size       uint64
	Type       string
	HeaderSize uint64
}

// ParseBoxHeader reads a box header from a byte slice at a given offset.
func ParseBoxHeader(data []byte, offset int) (*BoxHeader, int, error) {
	var err error
	var size32 uint32
	var boxType string

	size32, offset, err = readUint32(data, offset)
	if err != nil {
		return nil, offset, err
	}

	boxType, offset, err = readString(data, offset, 4)
	if err != nil {
		return nil, offset, err
	}

	h := &BoxHeader{Type: boxType, HeaderSize: 8}

	if size32 == 1 {
		h.Size, offset, err = readUint64(data, offset)
		if err != nil {
			return nil, offset, err
		}
		h.HeaderSize = 16
	} else {
		h.Size = uint64(size32)
	}

	return h, offset, nil
}