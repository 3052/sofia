package sofia

import (
   "encoding/binary"
   "io"
)

type RunSample struct {
   Duration uint32
   Size uint32
   Flags uint32
   CompositionTimeOffset [4]byte
}

func (s *RunSample) Decode(r io.Reader, t *TrackRun) error {
   if t.SampleDurationPresent() {
      err := binary.Read(r, binary.BigEndian, &s.Duration)
      if err != nil {
         return err
      }
   }
   if t.SampleSizePresent() {
      err := binary.Read(r, binary.BigEndian, &s.Size)
      if err != nil {
         return err
      }
   }
   if t.SampleFlagsPresent() {
      err := binary.Read(r, binary.BigEndian, &s.Flags)
      if err != nil {
         return err
      }
   }
   if t.SampleCompositionTimeOffsetsPresent() {
      _, err := io.ReadFull(r, s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (s RunSample) Encode(w io.Writer, t TrackRun) error {
   if t.SampleDurationPresent() {
      err := binary.Write(w, binary.BigEndian, s.Duration)
      if err != nil {
         return err
      }
   }
   if t.SampleSizePresent() {
      err := binary.Write(w, binary.BigEndian, s.Size)
      if err != nil {
         return err
      }
   }
   if t.SampleFlagsPresent() {
      err := binary.Write(w, binary.BigEndian, s.Flags)
      if err != nil {
         return err
      }
   }
   if t.SampleCompositionTimeOffsetsPresent() {
      _, err := w.Write(s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

// ISO/IEC 14496-12
//
// If the data-offset is present, it is relative to the base-data-offset
// established in the track fragment header.
//
//  aligned(8) class TrackRunBox extends FullBox('trun', version, tr_flags) {
//     unsigned int(32) sample_count;
//     signed int(32) data_offset; // 0x000001, assume present
//     unsigned int(32) first_sample_flags; // 0x000004
//     {
//        unsigned int(32) sample_duration; // 0x000100
//        unsigned int(32) sample_size; // 0x000200, assume present
//        unsigned int(32) sample_flags // 0x000400
//        if (version == 0) {
//           unsigned int(32) sample_composition_time_offset; // 0x000800
//        } else {
//           signed int(32) sample_composition_time_offset; // 0x000800
//        }
//     }[ sample_count ]
//  }
type TrackRun struct {
   BoxHeader          BoxHeader
   FullBoxHeader      FullBoxHeader
   SampleCount       uint32
   DataOffset        int32
   FirstSampleFlags uint32
   Sample            []RunSample
}

func (t *TrackRun) Decode(r io.Reader) error {
   err := t.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &t.SampleCount)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &t.DataOffset)
   if err != nil {
      return err
   }
   if t.FirstSampleFlagsPresent() {
      err := binary.Read(r, binary.BigEndian, &t.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   t.Sample = make([]RunSample, t.SampleCount)
   for i, sample := range t.Sample {
      err := sample.Decode(r, t)
      if err != nil {
         return err
      }
      t.Sample[i] = sample
   }
   return nil
}

func (t TrackRun) Encode(w io.Writer) error {
   err := t.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := t.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, t.SampleCount); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, t.DataOffset); err != nil {
      return err
   }
   if t.FirstSampleFlagsPresent() {
      err := binary.Write(w, binary.BigEndian, t.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   for _, sample := range t.Sample {
      err := sample.Encode(w, t)
      if err != nil {
         return err
      }
   }
   return nil
}

// 0x000004 first-sample-flags-present
func (t TrackRun) FirstSampleFlagsPresent() bool {
   return t.FullBoxHeader.Flags()&4 >= 1
}

// 0x000800 sample-composition-time-offsets-present
func (t TrackRun) SampleCompositionTimeOffsetsPresent() bool {
   return t.FullBoxHeader.Flags()&0x800 >= 1
}

// 0x000100 sample-duration-present
func (t TrackRun) SampleDurationPresent() bool {
   return t.FullBoxHeader.Flags()&0x100 >= 1
}

// 0x000400 sample-flags-present
func (t TrackRun) SampleFlagsPresent() bool {
   return t.FullBoxHeader.Flags()&0x400 >= 1
}

// 0x000200 sample-size-present
func (t TrackRun) SampleSizePresent() bool {
   return t.FullBoxHeader.Flags() & 0x200 >= 1
}
