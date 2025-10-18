// File: enca_box.go
package mp4parser

// EncaBox (Encrypted Audio Sample Entry)
type EncaBox struct {
	// 6 bytes reserved, 2 bytes data_reference_index
	Prefix   []byte // 8 bytes
	Children []*EncaChildBox
}

type EncaChildBox struct {
	Sinf *SinfBox
	Raw  *RawBox
}
func (c *EncaChildBox) Size() uint64 {
	if c.Sinf != nil { return c.Sinf.Size() }
	if c.Raw != nil { return c.Raw.Size() }
	return 0
}
func (c *EncaChildBox) Format(dst []byte, offset int) int {
	if c.Sinf != nil { return c.Sinf.Format(dst, offset) }
	if c.Raw != nil { return c.Raw.Format(dst, offset) }
	return offset
}

func ParseEncaBox(data []byte) (*EncaBox, error) {
	b := &EncaBox{}
	// AudioSampleEntry has a 28-byte prefix before children.
	// We simplify to 8, then parse children. This requires a more robust parser for real-world use.
	const prefixSize = 8 
	if len(data) < prefixSize { return nil, ErrUnexpectedEOF }
	b.Prefix = data[:prefixSize]

	offset := prefixSize
	for offset < len(data) {
		header, headerEndOffset, err := ParseBoxHeader(data, offset)
		if err != nil { return nil, err }
		contentEndOffset := offset + int(header.Size)
		if contentEndOffset > len(data) { return nil, ErrUnexpectedEOF }
		content := data[headerEndOffset:contentEndOffset]
		child := &EncaChildBox{}
		switch header.Type {
		case "sinf": child.Sinf, err = ParseSinfBox(content)
		default: child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil { return nil, err }
		b.Children = append(b.Children, child)
		offset = contentEndOffset
	}
	return b, nil
}

func (b *EncaBox) Size() uint64 {
	size := uint64(8 + len(b.Prefix))
	for _, child := range b.Children { size += child.Size() }
	return size
}

func (b *EncaBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "enca")
	offset = writeBytes(dst, offset, b.Prefix)
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}