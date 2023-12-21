package sofia

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
   Samples []TrackRun
}

type TrackRun struct {
   Sample_Duration uint32
   Sample_Size uint32
   Sample_Flags uint32
   Sample_Composition_Time_Offset [4]byte
}
