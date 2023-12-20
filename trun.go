package sofia

// 0x000004 first-sample-flags-present; this overrides the default flags for the
// first sample only, defined in 8.8.3.1. This makes it possible to record a
// group of frames where the first is a key and the rest are difference frames,
// without supplying explicit flags for every sample. If this flag and field
// are used, sample-flags-present shall not be set.
// 
// aligned(8) class TrackRunBox extends FullBox(
//    'trun',
//    version,
//    tr_flags
// ) {
//    unsigned int(32) sample_count;
//    signed int(32) data_offset;
//    // the following are optional fields
//    unsigned int(32) first_sample_flags;
//    // all fields in the following array are optional
//    // as indicated by bits set in the tr_flags
//    {
//       unsigned int(32) sample_duration;
//       unsigned int(32) sample_size;
//       unsigned int(32) sample_flags
//       if (version == 0) {
//          unsigned int(32) sample_composition_time_offset;
//       } else {
//          signed int(32) sample_composition_time_offset;
//       }
//    }[ sample_count ]
// }
type TrackRunBox struct {
   Sample_Count uint32
}
