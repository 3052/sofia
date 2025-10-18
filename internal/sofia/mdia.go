// File: mdia_box.go
package mp4parser

type MdiaChildBox struct {
	Mdhd *MdhdBox
	Minf *MinfBox
	Raw  *RawBox
}
func (c *MdiaChildBox) Size() uint64 {
	switch {
	case c.Mdhd != nil: return c.Mdhd.Size()
	case c.Minf != nil: return c.Minf.Size()
	case c.Raw != nil: return c.Raw.Size()
	}
	return 0
}
func (c *MdiaChildBox) Format(dst []byte, offset int) int {
	switch {
	case c.Mdhd != nil: return c.Mdhd.Format(dst, offset)
	case c.Minf != nil: return c.Minf.Format(dst, offset)
	case c.Raw != nil: return c.Raw.Format(dst, offset)
	}
	return offset
}

type MdiaBox struct{ Children []*MdiaChildBox }
func ParseMdiaBox(data []byte) (*MdiaBox, error) {
	b := &MdiaBox{}
	offset := 0
	for offset < len(data) {
		header, headerEndOffset, err := ParseBoxHeader(data, offset)
		if err != nil { return nil, err }
		contentEndOffset := offset + int(header.Size)
		if contentEndOffset > len(data) { return nil, ErrUnexpectedEOF }
		content := data[headerEndOffset:contentEndOffset]
		child := &MdiaChildBox{}
		switch header.Type {
		case "mdhd": child.Mdhd, err = ParseMdhdBox(content)
		case "minf": child.Minf, err = ParseMinfBox(content)
		default: child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil { return nil, err }
		b.Children = append(b.Children, child)
		offset = contentEndOffset
	}
	return b, nil
}
func (b *MdiaBox) Size() uint64 {
	size := uint64(8)
	for _, child := range b.Children { size += child.Size() }
	return size
}
func (b *MdiaBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "mdia")
	for _, child := range b.Children { offset = child.Format(dst, offset) }
	return offset
}