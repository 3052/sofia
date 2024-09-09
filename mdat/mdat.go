package mdat

import (
   "154.pages.dev/sofia"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaDataBox extends Box('mdat') {
//      bit(8) data[];
//   }
type Box struct {
   Box sofia.Box
}

func (b *Box) Data(track TrackFragment) [][]byte {
   split := make([][]byte, track.TrackRun.SampleCount)
   for i := range split {
      size := max(
         track.TrackRun.Sample[i].SampleSize,
         track.FragmentHeader.DefaultSampleSize,
      )
      split[i] = b.Box.Payload[:size]
      b.Box.Payload = b.Box.Payload[size:]
   }
   return split
}

func (b *Box) read(src io.Reader) error {
   return b.Box.Read(src)
}

func (b *Box) write(dst io.Writer) error {
   return b.Box.Write(dst)
}
