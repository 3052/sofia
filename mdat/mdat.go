package mdat

import (
   "154.pages.dev/sofia"
   "154.pages.dev/sofia/traf"
   "io"
)

// BE CAREFUL WITH THE RECEIVER
func (b Box) Data(track traf.Box) [][]byte {
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

func (b *Box) Write(dst io.Writer) error {
   return b.Box.Write(dst)
}

// ISO/IEC 14496-12
//   aligned(8) class MediaDataBox extends Box('mdat') {
//      bit(8) data[];
//   }
type Box struct {
   Box sofia.Box
}

func (b *Box) Read(src io.Reader) error {
   return b.Box.Read(src)
}
