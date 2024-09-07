package sofia

import (
	"154.pages.dev/sofia/box"
	"io"
)

func (m *MediaInformation) read(r io.Reader, size int64) error {
	r = io.LimitReader(r, size)
	for {
		var head box.Header
		err := head.Read(r)
		switch err {
		case nil:
			switch head.Type.String() {
			case "stbl":
				_, size := head.get_size()
				m.SampleTable.BoxHeader = head
				err := m.SampleTable.read(r, size)
				if err != nil {
					return err
				}
			case "dinf", // Roku
				"smhd", // Roku
				"vmhd": // Roku
				object := box.Box{BoxHeader: head}
				err := object.read(r)
				if err != nil {
					return err
				}
				m.Boxes = append(m.Boxes, object)
			default:
				return box.Error{m.BoxHeader.Type, head.Type}
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

// ISO/IEC 14496-12
//
//	aligned(8) class MediaInformationBox extends Box('minf') {
//	}
type MediaInformation struct {
	BoxHeader   box.Header
	Boxes       []box.Box
	SampleTable SampleTable
}

func (m *MediaInformation) write(w io.Writer) error {
	err := m.BoxHeader.write(w)
	if err != nil {
		return err
	}
	for _, object := range m.Boxes {
		err := object.write(w)
		if err != nil {
			return err
		}
	}
	return m.SampleTable.write(w)
}
