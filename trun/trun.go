package trun

import (
   "154.pages.dev/sofia"
   "encoding/binary"
)

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

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf = binary.BigEndian.AppendUint32(buf, b.SampleCount)
   buf, err = binary.Append(buf, binary.BigEndian, b.DataOffset)
   if err != nil {
      return nil, err
   }
   if b.first_sample_flags_present() {
      buf = binary.BigEndian.AppendUint32(buf, b.FirstSampleFlags)
   }
   for _, value := range b.Sample {
      buf, err = value.Append(buf)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
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

// 0x000800 sample-composition-time-offsets-present
func (b *Box) sample_composition_time_offsets_present() bool {
   return b.FullBoxHeader.GetFlags()&0x800 >= 1
}

func (b *Box) Read(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.DataOffset)
   if err != nil {
      return err
   }
   buf = buf[n:]
   if b.first_sample_flags_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.FirstSampleFlags)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   b.Sample = make([]Sample, b.SampleCount)
   for i, value := range b.Sample {
      value.box = b
      n, err = value.Decode(buf)
      if err != nil {
         return err
      }
      buf = buf[n:]
      b.Sample[i] = value
   }
   return nil
}

func (s *Sample) Decode(buf []byte) (int, error) {
   var ns int
   if s.box.sample_duration_present() {
      n, err := binary.Decode(buf[ns:], binary.BigEndian, &s.Duration)
      if err != nil {
         return 0, err
      }
      ns += n
   }
   if s.box.sample_size_present() {
      n, err := binary.Decode(buf[ns:], binary.BigEndian, &s.SampleSize)
      if err != nil {
         return 0, err
      }
      ns += n
   }
   if s.box.sample_flags_present() {
      n, err := binary.Decode(buf[ns:], binary.BigEndian, &s.Flags)
      if err != nil {
         return 0, err
      }
      ns += n
   }
   if s.box.sample_composition_time_offsets_present() {
      ns += copy(s.CompositionTimeOffset[:], buf[ns:])
   }
   return ns, nil
}

func (s *Sample) Append(buf []byte) ([]byte, error) {
   if s.box.sample_duration_present() {
      buf = binary.BigEndian.AppendUint32(buf, s.Duration)
   }
   if s.box.sample_size_present() {
      buf = binary.BigEndian.AppendUint32(buf, s.SampleSize)
   }
   if s.box.sample_flags_present() {
      buf = binary.BigEndian.AppendUint32(buf, s.Flags)
   }
   if s.box.sample_composition_time_offsets_present() {
      buf = append(buf, s.CompositionTimeOffset[:]...)
   }
   return buf, nil
}

type Sample struct {
   Duration              uint32
   SampleSize            uint32
   Flags                 uint32
   CompositionTimeOffset [4]byte
   box                   *Box
}
