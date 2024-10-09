package mdat

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/traf"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaDataBox extends Box('mdat') {
//      bit(8) data[];
//   }
type Box struct {
   Box sofia.Box
}

// BE CAREFUL WITH THE RECEIVER
func (b *Box) Data(track *traf.Box) [][]byte {
   payload := b.Box.Payload
   data := make([][]byte, track.Trun.SampleCount)
   for i, s := range track.Trun.Sample {
      if s.SampleSize == 0 {
         s.SampleSize = track.Tfhd.DefaultSampleSize
      }
      data[i] = payload[:s.SampleSize]
      payload = payload[s.SampleSize:]
   }
   return data
}
