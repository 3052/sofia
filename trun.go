package sofia

import (
   "encoding/binary"
   "io"
)

// If the data-offset is present, it is relative to the base-data-offset
// established in the track fragment header.
// 
// aligned(8) class TrackRunBox extends FullBox(
//    'trun',
//    version,
//    tr_flags
// ) {
//    unsigned int(32) sample_count;
//    signed int(32) data_offset; // 0x000001, assume present
//    unsigned int(32) first_sample_flags; // 0x000004
//    {
//       unsigned int(32) sample_duration; // 0x000100
//       unsigned int(32) sample_size; // 0x000200, assume present
//       unsigned int(32) sample_flags // 0x000400
//       if (version == 0) {
//          unsigned int(32) sample_composition_time_offset; // 0x000800
//       } else {
//          signed int(32) sample_composition_time_offset; // 0x000800
//       }
//    }[ sample_count ]
// }
type TrackRunBox struct {
   Header FullBoxHeader
   Sample_Count uint32
   Data_Offset int32
   First_Sample_Flags uint32
   Samples []SampleTrackRun
}

type SampleTrackRun struct {
   Duration uint32
   Size uint32
   Flags uint32
   Composition_Time_Offset [4]byte
}

func (t TrackRunBox) Composition_Time_Offsets_Present() bool {
   return t.Header.Flags & 0x800 >= 1
}

func (t TrackRunBox) Duration_Present() bool {
   return t.Header.Flags & 0x100 >= 1
}

func (t TrackRunBox) Flags_Present() bool {
   return t.Header.Flags & 0x400 >= 1
}

func (t TrackRunBox) Sample_Flags_Present() bool {
   return t.Header.Flags & 4 >= 1
}

func (t *TrackRunBox) Decode(r io.Reader) error {
   err := t.Header.Decode(r)
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
   if t.Sample_Flags_Present() {
      err := binary.Read(r, binary.BigEndian, &t.First_Sample_Flags)
      if err != nil {
         return err
      }
   }
   for count := t.Sample_Count; count >= 1; count-- {
      var run SampleTrackRun
      if t.Duration_Present() {
         err := binary.Read(r, binary.BigEndian, &run.Duration)
         if err != nil {
            return err
         }
      }
      err := binary.Read(r, binary.BigEndian, &run.Size)
      if err != nil {
         return err
      }
      if t.Flags_Present() {
         err := binary.Read(r, binary.BigEndian, &run.Flags)
         if err != nil {
            return err
         }
      }
      if t.Composition_Time_Offsets_Present() {
         _, err := r.Read(run.Composition_Time_Offset[:])
         if err != nil {
            return err
         }
      }
      t.Samples = append(t.Samples, run)
   }
   return nil
}
