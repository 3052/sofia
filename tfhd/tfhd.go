package tfhd

import (
   "154.pages.dev/sofia"
   "encoding/binary"
   "io"
)

func (b *Box) Read(src io.Reader) error {
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   err = binary.Read(src, binary.BigEndian, &b.TrackId)
   if err != nil {
      return err
   }
   if b.sample_description_index_present() {
      err := binary.Read(src, binary.BigEndian, &b.SampleDescriptionIndex)
      if err != nil {
         return err
      }
   }
   if b.default_sample_duration_present() {
      err := binary.Read(src, binary.BigEndian, &b.DefaultSampleDuration)
      if err != nil {
         return err
      }
   }
   if b.default_sample_size_present() {
      err := binary.Read(src, binary.BigEndian, &b.DefaultSampleSize)
      if err != nil {
         return err
      }
   }
   if b.default_sample_flags_present() {
      err := binary.Read(src, binary.BigEndian, &b.DefaultSampleFlags)
      if err != nil {
         return err
      }
   }
   return nil
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = binary.Write(dst, binary.BigEndian, b.TrackId)
   if err != nil {
      return err
   }
   if b.sample_description_index_present() {
      err := binary.Write(dst, binary.BigEndian, b.SampleDescriptionIndex)
      if err != nil {
         return err
      }
   }
   if b.default_sample_duration_present() {
      err := binary.Write(dst, binary.BigEndian, b.DefaultSampleDuration)
      if err != nil {
         return err
      }
   }
   if b.default_sample_size_present() {
      err := binary.Write(dst, binary.BigEndian, b.DefaultSampleSize)
      if err != nil {
         return err
      }
   }
   if b.default_sample_flags_present() {
      err := binary.Write(dst, binary.BigEndian, b.DefaultSampleFlags)
      if err != nil {
         return err
      }
   }
   return nil
}

// 0x000002 sample-description-index-present
func (b *Box) sample_description_index_present() bool {
   return b.FullBoxHeader.GetFlags()&0x2 >= 1
}

// 0x000008 default-sample-duration-present
func (b *Box) default_sample_duration_present() bool {
   return b.FullBoxHeader.GetFlags()&0x8 >= 1
}

// 0x000010 default-sample-size-present
func (b *Box) default_sample_size_present() bool {
   return b.FullBoxHeader.GetFlags()&0x10 >= 1
}

// 0x000020 default-sample-flags-present
func (b *Box) default_sample_flags_present() bool {
   return b.FullBoxHeader.GetFlags()&0x20 >= 1
}

// ISO/IEC 14496-12
//   aligned(8) class TrackFragmentHeaderBox extends FullBox(
//      'tfhd', 0, tf_flags
//   ) {
//      unsigned int(32) track_ID;
//      // all the following are optional fields
//      // their presence is indicated by bits in the tf_flags
//      unsigned int(64) base_data_offset; // ASSUME NOT PRESENT
//      unsigned int(32) sample_description_index;
//      unsigned int(32) default_sample_duration;
//      unsigned int(32) default_sample_size;
//      unsigned int(32) default_sample_flags;
//   }
type Box struct {
   BoxHeader              sofia.BoxHeader
   FullBoxHeader          sofia.FullBoxHeader
   TrackId                uint32
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}
