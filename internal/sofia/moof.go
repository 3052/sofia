// File: moof_box.go
package mp4parser

// MoofBox (Movie Fragment Box)
type MoofBox struct {
	Mfhd *MfhdBox
	Traf []*TrafBox
}

// ParseMoofBox parses the MoofBox from its content slice.
func ParseMoofBox(data []byte) (*MoofBox, error) {
	b := &MoofBox{}
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
		switch header.Type {
		case "mfhd":
			b.Mfhd, err = ParseMfhdBox(content)
		case "traf":
			var traf *TrafBox
			traf, err = ParseTrafBox(content)
			b.Traf = append(b.Traf, traf)
		}
		if err != nil {
			return nil, err
		}
		offset = contentEndOffset
	}
	return b, nil
}

// Size calculates the total byte size of the MoofBox.
func (b *MoofBox) Size() uint64 {
	size := uint64(8) // Header size
	if b.Mfhd != nil {
		size += b.Mfhd.Size()
	}
	for _, traf := range b.Traf {
		size += traf.Size()
	}
	return size
}

// Format serializes the MoofBox into the destination slice and returns the new offset.
func (b *MoofBox) Format(dst []byte, offset int) int {
	offset = writeUint32(dst, offset, uint32(b.Size()))
	offset = writeString(dst, offset, "moof")
	if b.Mfhd != nil {
		offset = b.Mfhd.Format(dst, offset)
	}
	for _, traf := range b.Traf {
		offset = traf.Format(dst, offset)
	}
	return offset
}