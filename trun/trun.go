package trun

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/mdhd"
   "41.neocities.org/sofia/tfhd"
   "encoding/binary"
)

func (s *Sample) Decode(box1 *Box, data []byte) (int, error) {
   var n int
   if box1.sample_duration_present() {
      n1, err := binary.Decode(data[n:], binary.BigEndian, &s.Duration)
      if err != nil {
         return 0, err
      }
      n += n1
   }
   if box1.sample_size_present() {
      n1, err := binary.Decode(data[n:], binary.BigEndian, &s.Size)
      if err != nil {
         return 0, err
      }
      n += n1
   }
   if box1.sample_flags_present() {
      n1, err := binary.Decode(data[n:], binary.BigEndian, &s.Flags)
      if err != nil {
         return 0, err
      }
      n += n1
   }
   if box1.sample_composition_time_offsets_present() {
      n += copy(s.CompositionTimeOffset[:], data[n:])
   }
   return n, nil
}

func (s *Sample) Append(box1 *Box, data []byte) ([]byte, error) {
   if box1.sample_duration_present() {
      data = binary.BigEndian.AppendUint32(data, s.Duration)
   }
   if box1.sample_size_present() {
      data = binary.BigEndian.AppendUint32(data, s.Size)
   }
   if box1.sample_flags_present() {
      data = binary.BigEndian.AppendUint32(data, s.Flags)
   }
   if box1.sample_composition_time_offsets_present() {
      data = append(data, s.CompositionTimeOffset[:]...)
   }
   return data, nil
}
func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.DataOffset)
   if err != nil {
      return err
   }
   data = data[n:]
   if b.first_sample_flags_present() {
      n, err = binary.Decode(data, binary.BigEndian, &b.FirstSampleFlags)
      if err != nil {
         return err
      }
      data = data[n:]
   }
   b.Sample = make([]Sample, 0, b.SampleCount)
   for _, sample1 := range b.Sample {
      n, err = sample1.Decode(b, data)
      if err != nil {
         return err
      }
      data = data[n:]
      b.Sample = append(b.Sample, sample1)
   }
   return nil
}

func (s *Sample) Bandwidth(track *tfhd.Box, media *mdhd.Box) uint64 {
   if s.Duration == 0 {
      s.Duration = track.DefaultSampleDuration
   }
   if s.Size == 0 {
      s.Size = track.DefaultSampleSize
   }
   return uint64(media.Timescale) * uint64(s.Size) * 8 / uint64(s.Duration)
}

type Sample struct {
   Duration              uint32
   Size                  uint32
   Flags                 uint32
   CompositionTimeOffset [4]byte
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, b.FullBoxHeader)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.SampleCount)
   data, err = binary.Append(data, binary.BigEndian, b.DataOffset)
   if err != nil {
      return nil, err
   }
   if b.first_sample_flags_present() {
      data = binary.BigEndian.AppendUint32(data, b.FirstSampleFlags)
   }
   for _, sample1 := range b.Sample {
      data, err = sample1.Append(b, data)
      if err != nil {
         return nil, err
      }
   }
   return data, nil
}

// ISO/IEC 14496-12
//
// If the data-offset is present, it is relative to the base-data-offset
// established in the track fragment header.
//
// sample-size-present: each sample has its own size, otherwise the default is
// used.
//
//   aligned(8) class TrackRunBox extends FullBox('trun', version, tr_flags) {
//      unsigned int(32) sample_count;
//      signed int(32) data_offset; // 0x000001, assume present
//      unsigned int(32) first_sample_flags; // 0x000004
//      {
//         unsigned int(32) sample_duration; // 0x000100
//         unsigned int(32) sample_size; // 0x000200
//         unsigned int(32) sample_flags // 0x000400
//         if (version == 0) {
//            unsigned int(32) sample_composition_time_offset; // 0x000800
//         } else {
//            signed int(32) sample_composition_time_offset; // 0x000800
//         }
//      }[ sample_count ]
//   }
type Box struct {
   BoxHeader        sofia.BoxHeader
   FullBoxHeader    sofia.FullBoxHeader
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Sample           []Sample
}

// 0x000800 sample-composition-time-offsets-present
func (b *Box) sample_composition_time_offsets_present() bool {
   return b.FullBoxHeader.GetFlags()&0x800 >= 1
}

// 0x000004 first-sample-flags-present
func (b *Box) first_sample_flags_present() bool {
   return b.FullBoxHeader.GetFlags()&0x4 >= 1
}

// 0x000100 sample-duration-present
func (b *Box) sample_duration_present() bool {
   return b.FullBoxHeader.GetFlags()&0x100 >= 1
}

// 0x000200 sample-size-present
func (b *Box) sample_size_present() bool {
   return b.FullBoxHeader.GetFlags()&0x200 >= 1
}

// 0x000400 sample-flags-present
func (b *Box) sample_flags_present() bool {
   return b.FullBoxHeader.GetFlags()&0x400 >= 1
}

