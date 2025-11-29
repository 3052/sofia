package sofia

import (
   "encoding/binary"
   "errors"
)

type TfhdBox struct {
   Header                 BoxHeader
   Flags                  uint32
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

func (b *TfhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   b.TrackID = binary.BigEndian.Uint32(data[12:16])
   offset := 16

   if b.Flags&0x000001 != 0 {
      if offset+8 > len(data) {
         return errors.New("tfhd too short")
      }
      b.BaseDataOffset = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   }
   if b.Flags&0x000002 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.SampleDescriptionIndex = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000008 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleDuration = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000010 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleSize = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   if b.Flags&0x000020 != 0 {
      if offset+4 > len(data) {
         return errors.New("tfhd too short")
      }
      b.DefaultSampleFlags = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
   }
   return nil
}
