package mp4parser

// TrafChildBox can hold any of the parsed child types or a raw box.
type TrafChildBox struct {
	Tfhd *TfhdBox
	Trun *TrunBox
	// Note: senc, tfdt, saiz, saio, free are not in the parse list, so they become RawBox
	Raw *RawBox
}

// Size calculates the size of the contained child.
func (c *TrafChildBox) Size() uint64 {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Size()
	case c.Trun != nil:
		return c.Trun.Size()
	case c.Raw != nil:
		return c.Raw.Size()
	}
	return 0
}

// Format formats the contained child.
func (c *TrafChildBox) Format(dst []byte, offset int) int {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Format(dst, offset)
	case c.Trun != nil:
		return c.Trun.Format(dst, offset)
	case c.Raw != nil:
		return c.Raw.Format(dst, offset)
	}
	return offset
}

// TrafBox (Track Fragment Box)
type TrafBox struct {
	Children []*TrafChildBox
}

// ParseTrafBox parses the TrafBox from its content slice.
func ParseTrafBox(data []byte) (*TrafBox, error) {
	b := &TrafBox{}
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
		child := &TrafChildBox{}
		switch header.Type {
		case "tfhd":
			child.Tfhd, err = ParseTfhdBox(content)
		case "trun":
			child.Trun, err = ParseTrunBox(content)
		// REMOVED: `case "senc"` now falls through to the default case.
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

// Size calculates the total byte size of the TrafBox.
func (b *TrafBox) Size() uint64 {
	size := uint64(8)
	for _, child := range b.Children {
		size += child.Size()
	}
	return size
}

// Format serializes the TrafBox into the destination slice.
func (b *TrafBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "traf")
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}
