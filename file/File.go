package file

import (
	"154.pages.dev/sofia"
	"io"
)

func (f *File) Read(r io.Reader) error {
	for {
		var head sofia.BoxHeader
		err := head.Read(r)
		switch err {
		case nil:
			_, size := head.GetSize()
			switch head.Type.String() {
			case "mdat":
				f.MediaData = &MediaData{}
				f.MediaData.Box.BoxHeader = head
				err := f.MediaData.read(r)
				if err != nil {
					return err
				}
			case "moof":
				f.MovieFragment = &MovieFragment{BoxHeader: head}
				err := f.MovieFragment.read(r, size)
				if err != nil {
					return err
				}
			case "sidx":
				f.SegmentIndex = &SegmentIndex{BoxHeader: head}
				err := f.SegmentIndex.read(r)
				if err != nil {
					return err
				}
			case "moov":
				f.Movie = &Movie{BoxHeader: head}
				err := f.Movie.read(r, size)
				if err != nil {
					return err
				}
			case "free", // Mubi
				"ftyp", // Roku
				"styp": // Roku
				object := sofia.Box{BoxHeader: head}
				err := object.Read(r)
				if err != nil {
					return err
				}
				f.Boxes = append(f.Boxes, object)
			default:
				var container box.Type
				copy(container[:], "File")
				return box.Error{container, head.Type}
			}
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

// ISO/IEC 14496-12
type File struct {
	Boxes         []sofia.Box
	MediaData     *MediaData
	Movie         *Movie
	MovieFragment *MovieFragment
	SegmentIndex  *SegmentIndex
}

func (f *File) GetMovie() (*Movie, bool) {
	if f.Movie != nil {
		return f.Movie, true
	}
	return nil, false
}

func (f *File) Write(w io.Writer) error {
	// KEEP THESE IN ORDER
	for _, value := range f.Boxes {
		err := value.Write(w)
		if err != nil {
			return err
		}
	}
	if f.Movie != nil { // moov
		err := f.Movie.write(w)
		if err != nil {
			return err
		}
	}
	if f.SegmentIndex != nil { // sidx
		err := f.SegmentIndex.write(w)
		if err != nil {
			return err
		}
	}
	if f.MovieFragment != nil { // moof
		err := f.MovieFragment.write(w)
		if err != nil {
			return err
		}
	}
	if f.MediaData != nil { // mdat
		err := f.MediaData.write(w)
		if err != nil {
			return err
		}
	}
	return nil
}
