package sofia

import (
	"errors"
	"io"
	"log/slog"
)

// ISO/IEC 14496-12
type File struct {
	Boxes         []Box
	MediaData     *MediaData
	Movie         *Movie
	MovieFragment *MovieFragment
	SegmentIndex  *SegmentIndex
}

func (f *File) Decode(r io.Reader) error {
	for {
		var head BoxHeader
		err := head.read(r)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		box_type := head.GetType()
		r := head.payload(r)
		slog.Debug("BoxHeader", "Type", box_type)
		switch box_type {
		case "free", // Mubi
			"ftyp", // Roku
			"styp": // Roku
			b := Box{BoxHeader: head}
			err := b.read(r)
			if err != nil {
				return err
			}
			f.Boxes = append(f.Boxes, b)
		case "mdat":
			f.MediaData = &MediaData{BoxHeader: head}
			err := f.MediaData.Decode(r, f.MovieFragment.TrackFragment.TrackRun)
			if err != nil {
				return err
			}
		case "moof":
			f.MovieFragment = &MovieFragment{BoxHeader: head}
			err := f.MovieFragment.Decode(r)
			if err != nil {
				return err
			}
		case "moov":
			f.Movie = &Movie{BoxHeader: head}
			err := f.Movie.Decode(r)
			if err != nil {
				return err
			}
		case "sidx":
			f.SegmentIndex = &SegmentIndex{BoxHeader: head}
			err := f.SegmentIndex.Decode(r)
			if err != nil {
				return err
			}
		default:
			return errors.New("File.Decode")
		}
	}
}

// KEEP THESE IN ORDER
func (f File) Encode(w io.Writer) error {
	for _, value := range f.Boxes {
		err := value.Encode(w)
		if err != nil {
			return err
		}
	}
	if f.Movie != nil { // moov
		err := f.Movie.Encode(w)
		if err != nil {
			return err
		}
	}
	if f.SegmentIndex != nil { // sidx
		err := f.SegmentIndex.Encode(w) // this might be optional
		if err != nil {
			return err
		}
	}
	if f.MovieFragment != nil { // moof
		err := f.MovieFragment.Encode(w)
		if err != nil {
			return err
		}
	}
	if f.MediaData != nil { // mdat
		err := f.MediaData.Encode(w)
		if err != nil {
			return err
		}
	}
	return nil
}
