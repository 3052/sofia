package sofia

import (
   "encoding/binary"
   "io"
)

// Container: TrackFragmentBox
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
type TrackRunBox struct {
   BoxHeader          BoxHeader
   FullBoxHeader      FullBoxHeader
   Sample_Count       uint32
   Data_Offset        int32
   First_Sample_Flags uint32
   Sample            []TrackRunSample
}

func (t *TrackRunBox) Decode(r io.Reader) error {
   err := t.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &t.Sample_Count)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &t.Data_Offset)
   if err != nil {
      return err
   }
   if t.First_Sample_Flags_Present() {
      err := binary.Read(r, binary.BigEndian, &t.First_Sample_Flags)
      if err != nil {
         return err
      }
   }
   t.Sample = make([]TrackRunSample, t.Sample_Count)
   for i, sample := range t.Sample {
      err := sample.Decode(r, t)
      if err != nil {
         return err
      }
      t.Sample[i] = sample
   }
   return nil
}

func (t TrackRunBox) Encode(w io.Writer) error {
   err := t.BoxHeader.Encode(w)
   if err != nil {
      return err
   }
   if err := t.FullBoxHeader.Encode(w); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, t.Sample_Count); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, t.Data_Offset); err != nil {
      return err
   }
   if t.First_Sample_Flags_Present() {
      err := binary.Write(w, binary.BigEndian, t.First_Sample_Flags)
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

func (s *TrackRunSample) Decode(r io.Reader, t *TrackRunBox) error {
   if t.Sample_Duration_Present() {
      err := binary.Read(r, binary.BigEndian, &s.Sample_Duration)
      if err != nil {
         return err
      }
   }
   if t.Sample_Size_Present() {
      err := binary.Read(r, binary.BigEndian, &s.Sample_Size)
      if err != nil {
         return err
      }
   }
   if t.Sample_Flags_Present() {
      err := binary.Read(r, binary.BigEndian, &s.Sample_Flags)
      if err != nil {
         return err
      }
   }
   if t.Sample_Composition_Time_Offsets_Present() {
      _, err := io.ReadFull(r, s.Sample_Composition_Time_Offset[:])
      if err != nil {
         return err
      }
   }
   return nil
}

// 0x000004 first-sample-flags-present
func (t TrackRunBox) First_Sample_Flags_Present() bool {
   return t.FullBoxHeader.Flags()&4 >= 1
}

// 0x000800 sample-composition-time-offsets-present
func (t TrackRunBox) Sample_Composition_Time_Offsets_Present() bool {
   return t.FullBoxHeader.Flags()&0x800 >= 1
}

// 0x000100 sample-duration-present
func (t TrackRunBox) Sample_Duration_Present() bool {
   return t.FullBoxHeader.Flags()&0x100 >= 1
}

// 0x000400 sample-flags-present
func (t TrackRunBox) Sample_Flags_Present() bool {
   return t.FullBoxHeader.Flags()&0x400 >= 1
}

// 0x000200 sample-size-present
func (t TrackRunBox) Sample_Size_Present() bool {
   return t.FullBoxHeader.Flags() & 0x200 >= 1
}

type TrackRunSample struct {
   Sample_Duration uint32
   Sample_Size uint32
   Sample_Flags uint32
   Sample_Composition_Time_Offset [4]byte
}

func (TrackRunSample) Size(t TrackRunBox) uint32 {
   var s uint32
   if t.Sample_Duration_Present() {
      s += 4
   }
   if t.Sample_Size_Present() {
      s += 4
   }
   if t.Sample_Flags_Present() {
      s += 4
   }
   if t.Sample_Composition_Time_Offsets_Present() {
      s += 4
   }
   return s
}

func (s TrackRunSample) Encode(w io.Writer, t TrackRunBox) error {
   if t.Sample_Duration_Present() {
      err := binary.Write(w, binary.BigEndian, s.Sample_Duration)
      if err != nil {
         return err
      }
   }
   if t.Sample_Size_Present() {
      err := binary.Write(w, binary.BigEndian, s.Sample_Size)
      if err != nil {
         return err
      }
   }
   if t.Sample_Flags_Present() {
      err := binary.Write(w, binary.BigEndian, s.Sample_Flags)
      if err != nil {
         return err
      }
   }
   if t.Sample_Composition_Time_Offsets_Present() {
      _, err := w.Write(s.Sample_Composition_Time_Offset[:])
      if err != nil {
         return err
      }
   }
   return nil
}
