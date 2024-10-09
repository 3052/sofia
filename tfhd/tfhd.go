package tfhd

import (
   "41.neocities.org/sofia"
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

// 0x000001 base-data-offset-present
func (b *Box) base_data_offset_present() bool {
   return b.FullBoxHeader.GetFlags()&0x1 >= 1
}

func (b *Box) Read(buf []byte) error {
   n, err := b.FullBoxHeader.Decode(buf)
   if err != nil {
      return err
   }
   buf = buf[n:]
   n, err = binary.Decode(buf, binary.BigEndian, &b.TrackId)
   if err != nil {
      return err
   }
   buf = buf[n:]
   if b.base_data_offset_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.BaseDataOffset)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   if b.sample_description_index_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.SampleDescriptionIndex)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   if b.default_sample_duration_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.DefaultSampleDuration)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   if b.default_sample_size_present() {
      n, err = binary.Decode(buf, binary.BigEndian, &b.DefaultSampleSize)
      if err != nil {
         return err
      }
      buf = buf[n:]
   }
   if b.default_sample_flags_present() {
      _, err = binary.Decode(buf, binary.BigEndian, &b.DefaultSampleFlags)
      if err != nil {
         return err
      }
   }
   return nil
}

// ISO/IEC 14496-12
//
//   aligned(8) class TrackFragmentHeaderBox extends FullBox(
//      'tfhd', 0, tf_flags
//   ) {
//      unsigned int(32) track_ID;
//      // all the following are optional fields
//      // their presence is indicated by bits in the tf_flags
//      unsigned int(64) base_data_offset;
//      unsigned int(32) sample_description_index;
//      unsigned int(32) default_sample_duration;
//      unsigned int(32) default_sample_size;
//      unsigned int(32) default_sample_flags;
//   }
type Box struct {
   BoxHeader              sofia.BoxHeader
   FullBoxHeader          sofia.FullBoxHeader
   TrackId                uint32
   BaseDataOffset         uint64
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
   buf = binary.BigEndian.AppendUint32(buf, b.TrackId)
   if b.base_data_offset_present() {
      buf = binary.BigEndian.AppendUint64(buf, b.BaseDataOffset)
   }
   if b.sample_description_index_present() {
      buf = binary.BigEndian.AppendUint32(buf, b.SampleDescriptionIndex)
   }
   if b.default_sample_duration_present() {
      buf = binary.BigEndian.AppendUint32(buf, b.DefaultSampleDuration)
   }
   if b.default_sample_size_present() {
      buf = binary.BigEndian.AppendUint32(buf, b.DefaultSampleSize)
   }
   if b.default_sample_flags_present() {
      buf = binary.BigEndian.AppendUint32(buf, b.DefaultSampleFlags)
   }
   return buf, nil
}
