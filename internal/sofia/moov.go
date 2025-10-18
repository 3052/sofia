package mp4parser

// MoovChildBox can hold any of the parsed child types or a raw box.
type MoovChildBox struct {
	Trak *TrakBox
	// Pssh is no longer a special type; it will be handled by RawBox.
	Raw *RawBox
}

// Size calculates the size of the contained child.
func (c *MoovChildBox) Size() uint64 {
	switch {
	case c.Trak != nil:
		return c.Trak.Size()
	case c.Raw != nil:
		return c.Raw.Size()
	}
	return 0
}

// Format formats the contained child.
func (c *MoovChildBox) Format(dst []byte, offset int) int {
	switch {
	case c.Trak != nil:
		return c.Trak.Format(dst, offset)
	case c.Raw != nil:
		return c.Raw.Format(dst, offset)
	}
	return offset
}

// MoovBox (Movie Box)
type MoovBox struct {
	Children []*MoovChildBox
}

// ParseMoovBox parses the MoovBox from its content slice.
func ParseMoovBox(data []byte) (*MoovBox, error) {
	b := &MoovBox{}
	offset := 0
	for offset < len(data) {
		header, headerEndOffset, err := ParseBoxHeader(data, offset)
		if err != nil {
			return nil, err
		}
		contentEndOffset := offset + int(header.Size)
		if contentEndOffset > len(data) {
			return nil, ErrUnexpectedEOF
		}
		content := data[headerEndOffset:contentEndOffset]
		child := &MoovChildBox{}
		switch header.Type {
		case "trak":
			child.Trak, err = ParseTrakBox(content)
		// REMOVED: case "pssh" - This now falls through to the default.
		default:
			child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil {
			return nil, err
		}
		b.Children = append(b.Children, child)
		offset = contentEndOffset
	}
	return b, nil
}

// Size calculates the total byte size of the MoovBox.
func (b *MoovBox) Size() uint64 {
	size := uint64(8)
	for _, child := range b.Children {
		size += child.Size()
	}
	return size
}

// Format serializes the MoovBox into the destination slice.
func (b *MoovBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "moov")
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}
