// File: box.go
package mp4parser

import "fmt"

// Box is a container for a parsed top-level MP4 box.
type Box struct {
	Header *BoxHeader
	Moov   *MoovBox
	Moof   *MoofBox
	Mdat   *MdatBox
	Sidx   *SidxBox
	Raw    *RawBox // For ftyp and any other non-parsed top-level boxes
}

// Size calculates the total size of the contained top-level box.
func (b *Box) Size() uint64 {
	switch {
	case b.Moov != nil:
		return b.Moov.Size()
	case b.Moof != nil:
		return b.Moof.Size()
	case b.Mdat != nil:
		return b.Mdat.Size()
	case b.Sidx != nil:
		return b.Sidx.Size()
	case b.Raw != nil:
		return b.Raw.Size()
	default:
		return 0
	}
}

// Format serializes the top-level box into a new byte slice.
func (b *Box) Format() ([]byte, error) {
	size := b.Size()
	if size == 0 {
		return nil, fmt.Errorf("box is empty, cannot format")
	}
	dst := make([]byte, size)
	switch {
	case b.Moov != nil:
		b.Moov.Format(dst, 0)
	case b.Moof != nil:
		b.Moof.Format(dst, 0)
	case b.Mdat != nil:
		b.Mdat.Format(dst, 0)
	case b.Sidx != nil:
		b.Sidx.Format(dst, 0)
	case b.Raw != nil:
		b.Raw.Format(dst, 0)
	}
	return dst, nil
}