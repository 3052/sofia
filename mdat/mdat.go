package mdat

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/traf"
)

// ISO/IEC 14496-12
//   aligned(8) class MediaDataBox extends Box('mdat') {
//      bit(8) data[];
//   }
type Box [1]sofia.Box

// BE CAREFUL WITH THE RECEIVER
func (b Box) Data(track *traf.Box) [][]byte {
   data := make([][]byte, 0, track.Trun.SampleCount)
   for _, sample := range track.Trun.Sample {
      if sample.Size == 0 {
         sample.Size = track.Tfhd.DefaultSampleSize
      }
      data = append(data, b[0].Payload[:sample.Size])
      b[0].Payload = b[0].Payload[sample.Size:]
   }
   return data
}
