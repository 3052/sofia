package file

import (
	"154.pages.dev/sofia"
	"io"
)

// ISO/IEC 14496-12
//
//	aligned(8) class MediaDataBox extends Box('mdat') {
//	   bit(8) data[];
//	}
type MediaData struct {
	Box sofia.Box
}

func (m *MediaData) Data(track TrackFragment) [][]byte {
	split := make([][]byte, track.TrackRun.SampleCount)
	for i := range split {
		size := max(
			track.TrackRun.Sample[i].SampleSize,
			track.FragmentHeader.DefaultSampleSize,
		)
		split[i] = m.Box.Payload[:size]
		m.Box.Payload = m.Box.Payload[size:]
	}
	return split
}

func (m *MediaData) read(r io.Reader) error {
	return m.Box.Read(r)
}

func (m *MediaData) write(w io.Writer) error {
	return m.Box.Write(w)
}
