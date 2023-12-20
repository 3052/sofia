package sofia

// aligned(8) class TrackRunBox extends FullBox(
//    'trun',
//    version,
//    tr_flags
// ) {
//    unsigned int(32) sample_count;
//    // the following are optional fields
//    signed int(32) data_offset;
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
