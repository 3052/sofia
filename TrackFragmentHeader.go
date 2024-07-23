package sofia

import "io"

// ISO/IEC 14496-12
//  aligned(8) class TrackFragmentHeaderBox extends FullBox(
//     'tfhd', 0, tf_flags
//  ) {
//     unsigned int(32) track_ID;
//     // all the following are optional fields
//     // their presence is indicated by bits in the tf_flags
//     unsigned int(64) base_data_offset;
//     unsigned int(32) sample_description_index;
//     unsigned int(32) default_sample_duration;
//     unsigned int(32) default_sample_size;
//     unsigned int(32) default_sample_flags;
//  }
type TrackFragmentHeader struct {
   BoxHeader     BoxHeader
   FullBoxHeader FullBoxHeader
}

func (t *TrackFragmentHeader) read(r io.Reader) error {
   err := t.FullBoxHeader.read(r)
   if err != nil {
      return err
   }
   return nil
}
