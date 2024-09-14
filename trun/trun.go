package trun

import (
   "154.pages.dev/sofia"
   "encoding/binary"
)

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

// 0x000800 sample-composition-time-offsets-present
func (b *Box) sample_composition_time_offsets_present() bool {
   return b.FullBoxHeader.GetFlags()&0x800 >= 1
}

type Sample struct {
   Duration              uint32
   SampleSize            uint32
   Flags                 uint32
   CompositionTimeOffset [4]byte
   box                   *Box
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

func (s *Sample) Append(buf []byte) ([]byte, error) {
   var err error
   if s.box.sample_duration_present() {
      buf, err = binary.Append(buf, binary.BigEndian, s.Duration)
      if err != nil {
         return nil, err
      }
   }
   if s.box.sample_size_present() {
      buf, err = binary.Append(buf, binary.BigEndian, s.SampleSize)
      if err != nil {
         return nil, err
      }
   }
   if s.box.sample_flags_present() {
      buf, err = binary.Append(buf, binary.BigEndian, s.Flags)
      if err != nil {
         return nil, err
      }
   }
   if s.box.sample_composition_time_offsets_present() {
      buf = append(buf, s.CompositionTimeOffset[:]...)
   }
   return buf, nil
}

func (s *Sample) Decode(buf []byte) ([]byte, error) {
   if s.box.sample_duration_present() {
      n, err := binary.Decode(buf, binary.BigEndian, &s.Duration)
      if err != nil {
         return nil, err
      }
      buf = buf[n:]
   }
   if s.box.sample_size_present() {
      n, err := binary.Decode(buf, binary.BigEndian, &s.SampleSize)
      if err != nil {
         return nil, err
      }
      buf = buf[n:]
   }
   if s.box.sample_flags_present() {
      n, err := binary.Decode(buf, binary.BigEndian, &s.Flags)
      if err != nil {
         return nil, err
      }
      buf = buf[n:]
   }
   if s.box.sample_composition_time_offsets_present() {
      n := copy(s.CompositionTimeOffset[:], buf)
      buf = buf[n:]
   }
   return buf, nil
}

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.SampleCount)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.DataOffset)
   if err != nil {
      return nil, err
   }
   if b.first_sample_flags_present() {
      buf, err = binary.Append(buf, binary.BigEndian, b.FirstSampleFlags)
      if err != nil {
         return nil, err
      }
   }
   for _, value := range b.Sample {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

func (b *Box) Decode(buf []byte) ([]byte, error) {
   buf, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return nil, err
   }
   n, err := binary.Decode(buf, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return nil, err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.DataOffset)
   if err != nil {
      return nil, err
   }
   buf = buf[n:]
   if b.first_sample_flags_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.FirstSampleFlags)
      if err != nil {
         return nil, err
      }
      buf = buf[n:]
   }
   b.Sample = make([]Sample, b.SampleCount)
   for i, value := range b.Sample {
      value.box = b
      buf, err = value.Decode(buf)
      if err != nil {
         return nil, err
      }
      b.Sample[i] = value
   }
   return buf, nil
}
