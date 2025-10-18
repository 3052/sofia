package mp4parser

// StsdChildBox now only specifically handles encv, as enca will be a RawBox.
type StsdChildBox struct {
	Encv *EncvBox
	Raw  *RawBox
}

// Size calculates the size of the contained child.
func (c *StsdChildBox) Size() uint64 {
	if c.Encv != nil {
		return c.Encv.Size()
	}
	if c.Raw != nil {
		return c.Raw.Size()
	}
	return 0
}

// Format formats the contained child.
func (c *StsdChildBox) Format(dst []byte, offset int) int {
	if c.Encv != nil {
		return c.Encv.Format(dst, offset)
	}
	if c.Raw != nil {
		return c.Raw.Format(dst, offset)
	}
	return offset
}

// StsdBox (Sample Description Box)
type StsdBox struct {
	FullBox
	EntryCount uint32
	Children   []*StsdChildBox
}

// ParseStsdBox parses the StsdBox from its content slice.
func ParseStsdBox(data []byte) (*StsdBox, error) {
	b := &StsdBox{}
	offset, err := b.FullBox.Parse(data, 0)
	if err != nil {
		return nil, err
	}
	b.EntryCount, offset, err = readUint32(data, offset)
	if err != nil {
		return nil, err
	}
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
		child := &StsdChildBox{}
		switch header.Type {
		case "encv":
			child.Encv, err = ParseEncvBox(content)
		// REMOVED: `case "enca"` now falls through to the default case.
		default:
			child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil {
			return nil, err
		}
		b.Children = append(b.Children, child)
		offset = contentEndOffset
	}
	return b, err
}

// Size calculates the total byte size of the StsdBox.
func (b *StsdBox) Size() uint64 {
	size := uint64(8) + b.FullBox.Size() + 4 // Header, FullBox, EntryCount
	for _, child := range b.Children {
		size += child.Size()
	}
	return size
}

// Format serializes the StsdBox into the destination slice.
func (b *StsdBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "stsd")
	offset = b.FullBox.Format(dst, offset)
	offset = writeUint32(dst, offset, b.EntryCount)
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}
