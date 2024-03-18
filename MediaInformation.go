package sofia

import (
	"errors"
	"io"
	"log/slog"
)

// ISO/IEC 14496-12
//
//	aligned(8) class MediaInformationBox extends Box('minf') {
//	}
type MediaInformation struct {
	BoxHeader   BoxHeader
	Boxes       []Box
	SampleTable SampleTable
}

func (m *MediaInformation) Decode(r io.Reader) error {
	for {
		var head BoxHeader
		err := head.Decode(r)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		box_type := head.GetType()
		r := head.payload(r)
		slog.Debug("BoxHeader", "Type", box_type)
		switch box_type {
		case "dinf", // Roku
			"smhd", // Roku
			"vmhd": // Roku
			b := Box{BoxHeader: head}
			err := b.Decode(r)
			if err != nil {
				return err
			}
			m.Boxes = append(m.Boxes, b)
		case "stbl":
			m.SampleTable.BoxHeader = head
			err := m.SampleTable.Decode(r)
			if err != nil {
				return err
			}
		default:
			return errors.New("MediaInformation.Decode")
		}
	}
}

func (m MediaInformation) Encode(w io.Writer) error {
	err := m.BoxHeader.Encode(w)
	if err != nil {
		return err
	}
	for _, b := range m.Boxes {
		err := b.Encode(w)
		if err != nil {
			return err
		}
	}
	return m.SampleTable.Encode(w)
}
