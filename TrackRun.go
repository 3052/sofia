package sofia

import (
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
//         unsigned int(32) sample_size; // 0x000200, assume present
//         unsigned int(32) sample_flags // 0x000400
//         if (version == 0) {
//            unsigned int(32) sample_composition_time_offset; // 0x000800
//         } else {
//            signed int(32) sample_composition_time_offset; // 0x000800
//         }
//      }[ sample_count ]
//   }
type TrackRun struct {
   BoxHeader        BoxHeader
   FullBoxHeader    FullBoxHeader
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Sample           []RunSample
}

type RunSample struct {
   Duration              uint32
   Size                  uint32
   Flags                 uint32
   CompositionTimeOffset [4]byte
}

func (s *RunSample) read(r io.Reader, run *TrackRun) error {
   if run.sample_duration_present() {
      err := binary.Read(r, binary.BigEndian, &s.Duration)
      if err != nil {
         return err
      }
   }
   if run.sample_size_present() {
      err := binary.Read(r, binary.BigEndian, &s.Size)
      if err != nil {
         return err
      }
   }
   if run.sample_flags_present() {
      err := binary.Read(r, binary.BigEndian, &s.Flags)
      if err != nil {
         return err
      }
   }
   if run.sample_composition_time_offsets_present() {
      _, err := io.ReadFull(r, s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

func (s RunSample) write(w io.Writer, run TrackRun) error {
   if run.sample_duration_present() {
      err := binary.Write(w, binary.BigEndian, s.Duration)
      if err != nil {
         return err
      }
   }
   if run.sample_size_present() {
      err := binary.Write(w, binary.BigEndian, s.Size)
      if err != nil {
         return err
      }
   }
   if run.sample_flags_present() {
      err := binary.Write(w, binary.BigEndian, s.Flags)
      if err != nil {
         return err
      }
   }
   if run.sample_composition_time_offsets_present() {
      _, err := w.Write(s.CompositionTimeOffset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

// 0x000004 first-sample-flags-present
func (t TrackRun) first_sample_flags_present() bool {
   return t.FullBoxHeader.get_flags()&4 >= 1
}

func (t *TrackRun) read(r io.Reader) error {
   err := t.FullBoxHeader.read(r)
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
   if t.first_sample_flags_present() {
      err := binary.Read(r, binary.BigEndian, &t.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   t.Sample = make([]RunSample, t.SampleCount)
   for i, sample := range t.Sample {
      err := sample.read(r, t)
      if err != nil {
         return err
      }
      t.Sample[i] = sample
   }
   return nil
}

// 0x000800 sample-composition-time-offsets-present
func (t TrackRun) sample_composition_time_offsets_present() bool {
   return t.FullBoxHeader.get_flags()&0x800 >= 1
}

// 0x000100 sample-duration-present
func (t TrackRun) sample_duration_present() bool {
   return t.FullBoxHeader.get_flags()&0x100 >= 1
}

// 0x000400 sample-flags-present
func (t TrackRun) sample_flags_present() bool {
   return t.FullBoxHeader.get_flags()&0x400 >= 1
}

// 0x000200 sample-size-present
func (t TrackRun) sample_size_present() bool {
   return t.FullBoxHeader.get_flags()&0x200 >= 1
}

func (t TrackRun) write(w io.Writer) error {
   err := t.BoxHeader.write(w)
   if err != nil {
      return err
   }
   err = t.FullBoxHeader.write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, t.SampleCount)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, t.DataOffset)
   if err != nil {
      return err
   }
   if t.first_sample_flags_present() {
      err := binary.Write(w, binary.BigEndian, t.FirstSampleFlags)
      if err != nil {
         return err
      }
   }
   for _, sample := range t.Sample {
      err := sample.write(w, t)
      if err != nil {
         return err
      }
   }
   return nil
}
