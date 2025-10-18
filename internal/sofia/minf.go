// File: minf_box.go
package mp4parser

type MinfChildBox struct {
	Stbl *StblBox
	Raw  *RawBox
}
func (c *MinfChildBox) Size() uint64 {
	if c.Stbl != nil { return c.Stbl.Size() }
	if c.Raw != nil { return c.Raw.Size() }
	return 0
}
func (c *MinfChildBox) Format(dst []byte, offset int) int {
	if c.Stbl != nil { return c.Stbl.Format(dst, offset) }
	if c.Raw != nil { return c.Raw.Format(dst, offset) }
	return offset
}

type MinfBox struct{ Children []*MinfChildBox }
func ParseMinfBox(data []byte) (*MinfBox, error) {
	b := &MinfBox{}
	offset := 0
	for offset < len(data) {
		header, headerEndOffset, err := ParseBoxHeader(data, offset)
		if err != nil { return nil, err }
		contentEndOffset := offset + int(header.Size)
		if contentEndOffset > len(data) { return nil, ErrUnexpectedEOF }
		content := data[headerEndOffset:contentEndOffset]
		child := &MinfChildBox{}
		switch header.Type {
		case "stbl": child.Stbl, err = ParseStblBox(content)
		default: child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil { return nil, err }
		b.Children = append(b.Children, child)
		offset = contentEndOffset
	}
	return b, nil
}
func (b *MinfBox) Size() uint64 {
	size := uint64(8)
	for _, child := range b.Children { size += child.Size() }
	return size
}
func (b *MinfBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "minf")
	for _, child := range b.Children { offset = child.Format(dst, offset) }
	return offset
}