package sofia

import (
	"errors"
	"io"
	"log/slog"
)

// ISO/IEC 14496-12
//
//	aligned(8) class TrackFragmentBox extends Box('traf') {
//	}
type TrackFragment struct {
	BoxHeader        BoxHeader
	Boxes            []Box
	SampleEncryption SampleEncryption
	TrackRun         TrackRun
}

func (t *TrackFragment) Decode(r io.Reader) error {
	for {
		var head BoxHeader
		err := head.read(r)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		slog.Debug("BoxHeader", "type", head.GetType())
		r := head.payload(r)
		switch head.GetType() {
		case "saio", // Roku
			"saiz", // Roku
			"sbgp", // Roku
			"sgpd", // Roku
			"tfdt", // Roku
			"tfhd": // Roku
			b := Box{BoxHeader: head}
			err := b.read(r)
			if err != nil {
				return err
			}
			t.Boxes = append(t.Boxes, b)
		case "senc":
			t.SampleEncryption.BoxHeader = head
			err := t.SampleEncryption.Decode(r)
			if err != nil {
				return err
			}
		case "uuid":
			decode := func() bool {
				if head.get_usertype() == "a2394f525a9b4f14a2446c427c648df4" {
					if t.SampleEncryption.SampleCount == 0 {
						return true
					}
				}
				return false
			}
			if decode() {
				t.SampleEncryption.BoxHeader = head
				err := t.SampleEncryption.Decode(r)
				if err != nil {
					return err
				}
			} else {
				b := Box{BoxHeader: head}
				err := b.read(r)
				if err != nil {
					return err
				}
				t.Boxes = append(t.Boxes, b)
			}
		case "trun":
			t.TrackRun.BoxHeader = head
			err := t.TrackRun.Decode(r)
			if err != nil {
				return err
			}
		default:
			return errors.New("TrackFragment.Decode")
		}
	}
}

func (t TrackFragment) Encode(w io.Writer) error {
	err := t.BoxHeader.write(w)
	if err != nil {
		return err
	}
	for _, b := range t.Boxes {
		err := b.write(w)
		if err != nil {
			return err
		}
	}
	if err := t.TrackRun.Encode(w); err != nil {
		return err
	}
	return t.SampleEncryption.Encode(w)
}
