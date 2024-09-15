package tfhd

import (
   "154.pages.dev/sofia"
   "encoding/binary"
)

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

func (b *Box) Append(buf []byte) ([]byte, error) {
   buf, err := b.BoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = b.FullBoxHeader.Append(buf)
   if err != nil {
      return nil, err
   }
   buf, err = binary.Append(buf, binary.BigEndian, b.TrackId)
   if err != nil {
      return nil, err
   }
   if b.sample_description_index_present() {
      buf, err = binary.Append(buf, binary.BigEndian, b.SampleDescriptionIndex)
      if err != nil {
         return nil, err
      }
   }
   if b.default_sample_duration_present() {
      buf, err = binary.Append(buf, binary.BigEndian, b.DefaultSampleDuration)
      if err != nil {
         return nil, err
      }
   }
   if b.default_sample_size_present() {
      buf, err = binary.Append(buf, binary.BigEndian, b.DefaultSampleSize)
      if err != nil {
         return nil, err
      }
   }
   if b.default_sample_flags_present() {
      buf, err = binary.Append(buf, binary.BigEndian, b.DefaultSampleFlags)
      if err != nil {
         return nil, err
      }
   }
   return buf, nil
}

func (b *Box) Read(buf []byte) error {
   ns, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   n, err := binary.Decode(buf[ns:], binary.BigEndian, &b.TrackId)
   if err != nil {
      return err
   }
   ns += n
   if b.sample_description_index_present() {
      n, err = binary.Decode(
         buf[ns:], binary.BigEndian, &b.SampleDescriptionIndex,
      )
      if err != nil {
         return err
      }
      ns += n
   }
   if b.default_sample_duration_present() {
      n, err = binary.Decode(
         buf[ns:], binary.BigEndian, &b.DefaultSampleDuration,
      )
      if err != nil {
         return err
      }
      ns += n
   }
   if b.default_sample_size_present() {
      n, err = binary.Decode(buf[ns:], binary.BigEndian, &b.DefaultSampleSize)
      if err != nil {
         return err
      }
      ns += n
   }
   if b.default_sample_flags_present() {
      _, err = binary.Decode(buf[ns:], binary.BigEndian, &b.DefaultSampleFlags)
      if err != nil {
         return err
      }
   }
   return nil
}
