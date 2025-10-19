// File: traf.go
package mp4parser

import "log" // Import the log package

type TrafChildBox struct {
	Tfhd *TfhdBox
	Trun *TrunBox
	Senc *RawBox
	Raw  *RawBox
}

func (c *TrafChildBox) Size() uint64 {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Size()
	case c.Trun != nil:
		return c.Trun.Size()
	case c.Senc != nil:
		return c.Senc.Size()
	case c.Raw != nil:
		return c.Raw.Size()
	}
	return 0
}

func (c *TrafChildBox) Format(dst []byte, offset int) int {
	switch {
	case c.Tfhd != nil:
		return c.Tfhd.Format(dst, offset)
	case c.Trun != nil:
		return c.Trun.Format(dst, offset)
	case c.Senc != nil:
		return c.Senc.Format(dst, offset)
	case c.Raw != nil:
		return c.Raw.Format(dst, offset)
	}
	return offset
}

type TrafBox struct{ Children []*TrafChildBox }

func ParseTrafBox(data []byte) (*TrafBox, error) {
	b := &TrafBox{}
	offset := 0
	log.Printf("[ParseTrafBox] Starting to parse 'traf' box with total data length: %d", len(data))
	for offset < len(data) {
		log.Printf("[ParseTrafBox] Current offset: %d", offset)
		header, headerEndOffset, err := ParseBoxHeader(data, offset)
		if err != nil {
			log.Printf("[ParseTrafBox] ERROR parsing box header at offset %d: %v", offset, err)
			return nil, err
		}

		// LOGGING: Log details of the discovered box header
		log.Printf("[ParseTrafBox] Found box '%s' with total size %d (header size %d)", header.Type, header.Size, header.HeaderSize)

		contentEndOffset := offset + int(header.Size)
		if contentEndOffset > len(data) {
			log.Printf("[ParseTrafBox] ERROR: Box '%s' size (%d) exceeds available data (%d)", header.Type, header.Size, len(data)-offset)
			return nil, ErrUnexpectedEOF
		}

		content := data[headerEndOffset:contentEndOffset]

		// LOGGING: Log the size of the content slice we are about to parse
		log.Printf("[ParseTrafBox] Extracted content for '%s'. Content length: %d", header.Type, len(content))

		child := &TrafChildBox{}
		switch header.Type {
		case "tfhd":
			child.Tfhd, err = ParseTfhdBox(content)
		case "trun":
			child.Trun, err = ParseTrunBox(content)
		case "senc":
			child.Senc, err = ParseRawBox(header.Type, content)
		default:
			// Also logging other raw boxes that are not explicitly handled
			log.Printf("[ParseTrafBox] Treating box '%s' as a raw box.", header.Type)
			child.Raw, err = ParseRawBox(header.Type, content)
		}
		if err != nil {
			log.Printf("[ParseTrafBox] ERROR parsing content of box '%s': %v", header.Type, err)
			return nil, err
		}
		b.Children = append(b.Children, child)
		offset = contentEndOffset
		log.Printf("[ParseTrafBox] Advanced offset to: %d", offset)
	}
	log.Printf("[ParseTrafBox] Successfully finished parsing 'traf' box.")
	return b, nil
}

// ... rest of the file (Size, Format functions) remains the same
func (b *TrafBox) Size() uint64 {
	size := uint64(8)
	for _, child := range b.Children {
		size += child.Size()
	}
	return size
}

func (b *TrafBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "traf")
	for _, child := range b.Children {
		offset = child.Format(dst, offset)
	}
	return offset
}
