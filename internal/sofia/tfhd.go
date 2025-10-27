package mp4

import (
   "encoding/binary"
   "fmt"
)

// TfhdBox now includes all possible fields from the header.
type TfhdBox struct {
   Header                 BoxHeader
   RawData                []byte
   Flags                  uint32
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

// ParseTfhd is now a full parser that respects all flags.
func ParseTfhd(data []byte) (TfhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TfhdBox{}, err
   }
   var tfhd TfhdBox
   tfhd.Header = header
   tfhd.RawData = data[:header.Size]
   tfhd.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   tfhd.TrackID = binary.BigEndian.Uint32(data[12:16])
   offset := 16

   if tfhd.Flags&0x000001 != 0 { // base_data_offset_present
      if offset+8 > len(data) {
         return TfhdBox{}, fmt.Errorf("tfhd box too short for base_data_offset")
      }
      tfhd.BaseDataOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }
   if tfhd.Flags&0x000002 != 0 { // sample_description_index_present
      if offset+4 > len(data) {
         return TfhdBox{}, fmt.Errorf("tfhd box too short for sample_description_index")
      }
      tfhd.SampleDescriptionIndex = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if tfhd.Flags&0x000008 != 0 { // default_sample_duration_present
      if offset+4 > len(data) {
         return TfhdBox{}, fmt.Errorf("tfhd box too short for default_sample_duration")
      }
      tfhd.DefaultSampleDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if tfhd.Flags&0x000010 != 0 { // default_sample_size_present
      if offset+4 > len(data) {
         return TfhdBox{}, fmt.Errorf("tfhd box too short for default_sample_size")
      }
      tfhd.DefaultSampleSize = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if tfhd.Flags&0x000020 != 0 { // default_sample_flags_present
      if offset+4 > len(data) {
         return TfhdBox{}, fmt.Errorf("tfhd box too short for default_sample_flags")
      }
      tfhd.DefaultSampleFlags = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }

   return tfhd, nil
}

func (b *TfhdBox) Encode() []byte {
   return b.RawData
}
