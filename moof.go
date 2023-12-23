package sofia

import (
	"fmt"
	"io"
)

// aligned(8) class MovieFragmentBox extends Box('moof') {
// }
type MovieFragmentBox struct {
	Header BoxHeader
	Boxes  []Box
	Traf   TrackFragmentBox
}

func (m MovieFragmentBox) Encode(dst io.Writer) error {
	err := m.Header.Encode(dst)
	if err != nil {
		return err
	}
	for _, b := range m.Boxes {
		err := b.Encode(dst)
		if err != nil {
			return err
		}
	}
	return m.Traf.Encode(dst)
}

func (m *MovieFragmentBox) Decode(src io.Reader) error {
	for {
		var head BoxHeader
		err := head.Decode(src)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		size := head.BoxPayload()
		switch head.Type() {
		case "traf":
			m.Traf.Header = head
			err := m.Traf.Decode(io.LimitReader(src, size))
			if err != nil {
				return err
			}
		case "mfhd", "pssh":
			b := Box{Header: head}
			b.Payload = make([]byte, size)
			_, err := src.Read(b.Payload)
			if err != nil {
				return err
			}
			m.Boxes = append(m.Boxes, b)
		default:
			return fmt.Errorf("%q", head.RawType)
		}
	}
}
