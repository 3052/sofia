package trun

import (
   "154.pages.dev/sofia"
   "encoding/binary"
   "io"
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

func (b Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.SampleCount)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.DataOffset)
   if err != nil {
      return err
   }
   if b.first_sample_flags_present() {
      err := binary.Write(dst, binary.BigEndian, b.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   for _, value := range b.Sample {
      err := value.write(dst, b)
      if err != nil {
         return err
      }
   }
   return nil
}

func (b *Box) Read(src io.Reader) error {
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.DataOffset)
   if err != nil {
      return err
   }
   if b.first_sample_flags_present() {
      err := binary.Read(src, binary.BigEndian, &b.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   b.Sample = make([]Sample, b.SampleCount)
   for i, value := range b.Sample {
      err := value.read(src, b)
      if err != nil {
         return err
      }
      b.Sample[i] = value
   }
   return nil
}

type Sample struct {
   Duration              uint32
   SampleSize            uint32
   Flags                 uint32
   CompositionTimeOffset [4]byte
}

// 0x000004 first-sample-flags-present
func (b Box) first_sample_flags_present() bool {
   return b.FullBoxHeader.GetFlags()&0x4 >= 1
}

// 0x000100 sample-duration-present
func (b Box) sample_duration_present() bool {
   return b.FullBoxHeader.GetFlags()&0x100 >= 1
}

// 0x000200 sample-size-present
func (b Box) sample_size_present() bool {
   return b.FullBoxHeader.GetFlags()&0x200 >= 1
}

// 0x000400 sample-flags-present
func (b Box) sample_flags_present() bool {
   return b.FullBoxHeader.GetFlags()&0x400 >= 1
}

// 0x000800 sample-composition-time-offsets-present
func (b Box) sample_composition_time_offsets_present() bool {
   return b.FullBoxHeader.GetFlags()&0x800 >= 1
}

func (s *Sample) read(src io.Reader, run *Box) error {
   if run.sample_duration_present() {
      err := binary.Read(src, binary.BigEndian, &s.Duration)
      if err != nil {
         return err
      }
   }
   if run.sample_size_present() {
      err := binary.Read(src, binary.BigEndian, &s.SampleSize)
      if err != nil {
         return err
      }
   }
   if run.sample_flags_present() {
      err := binary.Read(src, binary.BigEndian, &s.Flags)
      if err != nil {
         return err
      }
   }
   if run.sample_composition_time_offsets_present() {
      _, err := io.ReadFull(src, s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (s Sample) write(dst io.Writer, run Box) error {
   if run.sample_duration_present() {
      err := binary.Write(dst, binary.BigEndian, s.Duration)
      if err != nil {
         return err
      }
   }
   if run.sample_size_present() {
      err := binary.Write(dst, binary.BigEndian, s.SampleSize)
      if err != nil {
         return err
      }
   }
   if run.sample_flags_present() {
      err := binary.Write(dst, binary.BigEndian, s.Flags)
      if err != nil {
         return err
      }
   }
   if run.sample_composition_time_offsets_present() {
      _, err := dst.Write(s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}
