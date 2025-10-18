package mp4parser

// TrafChildBox is a container for any box that can be a child of a TrafBox.
// This preserves the original order of the children.
type TrafChildBox struct {
	// Pointers to all possible child boxes. Only one will be non-nil.
	Tfhd *TfhdBox
	Tfdt *TfdtBox
	Senc *SencBox
	Saiz *SaizBox
	Saio *SaioBox
	Trun *TrunBox
	Free *FreeBox
}

// Size calculates the total byte size of the contained child box.
func (c *TrafChildBox) Size() uint64 {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Size()
	case c.Tfdt != nil:
		return c.Tfdt.Size()
	case c.Senc != nil:
		return c.Senc.Size()
	case c.Saiz != nil:
		return c.Saiz.Size()
	case c.Saio != nil:
		return c.Saio.Size()
	case c.Trun != nil:
		return c.Trun.Size()
	case c.Free != nil:
		return c.Free.Size()
	default:
		return 0
	}
}

// Format serializes the contained child box into the destination slice.
func (c *TrafChildBox) Format(dst []byte, offset int) int {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Format(dst, offset)
	case c.Tfdt != nil:
		return c.Tfdt.Format(dst, offset)
	case c.Senc != nil:
		return c.Senc.Format(dst, offset)
	case c.Saiz != nil:
		return c.Saiz.Format(dst, offset)
	case c.Saio != nil:
		return c.Saio.Format(dst, offset)
	case c.Trun != nil:
		return c.Trun.Format(dst, offset)
	case c.Free != nil:
		return c.Free.Format(dst, offset)
	default:
		return offset
	}
}

// TrafBox (Track Fragment Box) now holds an ordered slice of its children.
type TrafBox struct {
	Children []*TrafChildBox
}

// ParseTrafBox parses the TrafBox from its content slice, preserving child order.
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
		case "tfdt":
			child.Tfdt, err = ParseTfdtBox(content)
		case "senc":
			child.Senc, err = ParseSencBox(content)
		case "saiz":
			child.Saiz, err = ParseSaizBox(content)
		case "saio":
			child.Saio, err = ParseSaioBox(content)
		case "trun":
			child.Trun, err = ParseTrunBox(content)
		case "free":
			child.Free, err = ParseFreeBox(content)
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
	size := uint64(8) // Header size
	for _, child := range b.Children {
		size += child.Size()
	}
	return size
}

// Format serializes the TrafBox into the destination slice, preserving child order.
func (b *TrafBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "traf")
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}
