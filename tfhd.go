package sofia

import (
   "encoding/binary"
   "io"
)

// 8.8.7 Track fragment header box
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
type TrackFragmentHeaderBox struct {
   BoxHeader          BoxHeader
   FullBoxHeader      FullBoxHeader
   Track_ID uint32
   Base_Data_Offset uint64
   Sample_Description_Index uint32
   Default_Sample_Duration uint32
   Default_Sample_Size uint32
   Default_Sample_Flags uint32
}

func (b *TrackFragmentHeaderBox) Decode(r io.Reader) error {
   err := b.FullBoxHeader.Decode(r)
   if err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &b.Track_ID); err != nil {
      return err
   }
   if b.Base_Data_Offset_Present() {
      err := binary.Read(r, binary.BigEndian, &b.Base_Data_Offset)
      if err != nil {
         return err
      }
   }
   if b.Sample_Description_Index_Present() {
      err := binary.Read(r, binary.BigEndian, &b.Sample_Description_Index)
      if err != nil {
         return err
      }
   }
   if b.Default_Sample_Duration_Present() {
      err := binary.Read(r, binary.BigEndian, &b.Default_Sample_Duration)
      if err != nil {
         return err
      }
   }
   if b.Default_Sample_Size_Present() {
      err := binary.Read(r, binary.BigEndian, &b.Default_Sample_Size)
      if err != nil {
         return err
      }
   }
   if b.Default_Sample_Flags_Present() {
      err := binary.Read(r, binary.BigEndian, &b.Default_Sample_Flags)
      if err != nil {
         return err
      }
   }
   return nil
}

// 0x000001 base-data-offset-present
func (b TrackFragmentHeaderBox) Base_Data_Offset_Present() bool {
   return b.FullBoxHeader.Flags() & 1 >= 1
}

// 0x000002 sample-description-index-present
func (b TrackFragmentHeaderBox) Sample_Description_Index_Present() bool {
   return b.FullBoxHeader.Flags() & 2 >= 1
}

// 0x000008 default-sample-duration-present
func (b TrackFragmentHeaderBox) Default_Sample_Duration_Present() bool {
   return b.FullBoxHeader.Flags() & 8 >= 1
}

// 0x000010 default-sample-size-present
func (b TrackFragmentHeaderBox) Default_Sample_Size_Present() bool {
   return b.FullBoxHeader.Flags() & 0x10 >= 1
}

// 0x000020 default-sample-flags-present
func (b TrackFragmentHeaderBox) Default_Sample_Flags_Present() bool {
   return b.FullBoxHeader.Flags() & 0x20 >= 1
}
