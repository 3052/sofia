package sofia

import "io"

// ISO/IEC 14496-12
//  aligned(8) class MediaDataBox extends Box('mdat') {
//     bit(8) data[];
//  }
type MediaData struct {
   Box Box
}

func (m MediaData) Data(run TrackRun) [][]byte {
   split := make([][]byte, run.SampleCount)
   for i := range split {
      if j := run.Sample[i].Size; j >= 1 {
         split[i] = m.Box.Payload[:j]
         m.Box.Payload = m.Box.Payload[j:]
      } else {
         split[i] = m.Box.Payload
         m.Box.Payload = nil
      }
   }
   return split
}

func (m *MediaData) read(r io.Reader) error {
   return m.Box.read(r)
}

func (m MediaData) write(w io.Writer) error {
   return m.Box.write(w)
}
