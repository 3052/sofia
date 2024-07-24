package sofia

import "io"

func (m MediaData) Data(run TrackRun, header TrackFragmentHeader) [][]byte {
   split := make([][]byte, run.SampleCount)
   for i := range split {
      size := max(run.Sample[i].SampleSize, header.DefaultSampleSize)
      split[i] = m.Box.Payload[:size]
      m.Box.Payload = m.Box.Payload[size:]
   }
   return split
}

// ISO/IEC 14496-12
//  aligned(8) class MediaDataBox extends Box('mdat') {
//     bit(8) data[];
//  }
type MediaData struct {
   Box Box
}

func (m *MediaData) read(r io.Reader) error {
   return m.Box.read(r)
}

func (m MediaData) write(w io.Writer) error {
   return m.Box.write(w)
}
