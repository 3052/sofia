package mdat

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
   "io"
)

func (b *Box) Data(track traf.Box) [][]byte {
   split := make([][]byte, track.TrackRun.SampleCount)
   for i := range split {
      size := track.TrackRun.Sample[i].SampleSize
      if size == 0 {
         size = track.FragmentHeader.DefaultSampleSize
      }
      split[i] = b.Box.Payload[:size]
      b.Box.Payload = b.Box.Payload[size:]
   }
   return split
}

// ISO/IEC 14496-12
//   aligned(8) class MediaDataBox extends Box('mdat') {
//      bit(8) data[];
//   }
type Box struct {
   Box sofia.Box
}

func (b *Box) read(src io.Reader) error {
   return b.Box.Read(src)
}

func (b *Box) write(dst io.Writer) error {
   return b.Box.Write(dst)
}
