// File: tfhd_box.go
package mp4parser

type TfhdBox struct {
   FullBox
   TrackID                uint32
   BaseDataOffset         uint64
   SampleDescriptionIndex uint32
   DefaultSampleDuration  uint32
   DefaultSampleSize      uint32
   DefaultSampleFlags     uint32
}

func ParseTfhdBox(data []byte) (*TfhdBox, error) {
   b := &TfhdBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   b.TrackID, offset, err = readUint32(data, offset)
   if err != nil {
      return nil, err
   }
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      b.BaseDataOffset, offset, err = readUint64(data, offset)
      if err != nil {
         return nil, err
      }
   }
   if flags&0x000002 != 0 {
      b.SampleDescriptionIndex, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
   }
   if flags&0x000008 != 0 {
      b.DefaultSampleDuration, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
   }
   if flags&0x000010 != 0 {
      b.DefaultSampleSize, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
   }
   if flags&0x000020 != 0 {
      b.DefaultSampleFlags, _, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
   }
   return b, nil
}
func (b *TfhdBox) Size() uint64 {
   size := uint64(8) + b.FullBox.Size() + 4
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      size += 8
   }
   if flags&0x000002 != 0 {
      size += 4
   }
   if flags&0x000008 != 0 {
      size += 4
   }
   if flags&0x000010 != 0 {
      size += 4
   }
   if flags&0x000020 != 0 {
      size += 4
   }
   return size
}
func (b *TfhdBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "tfhd")
   offset = b.FullBox.Format(dst, offset)
   offset = writeUint32(dst, offset, b.TrackID)
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      offset = writeUint64(dst, offset, b.BaseDataOffset)
   }
   if flags&0x000002 != 0 {
      offset = writeUint32(dst, offset, b.SampleDescriptionIndex)
   }
   if flags&0x000008 != 0 {
      offset = writeUint32(dst, offset, b.DefaultSampleDuration)
   }
   if flags&0x000010 != 0 {
      offset = writeUint32(dst, offset, b.DefaultSampleSize)
   }
   if flags&0x000020 != 0 {
      offset = writeUint32(dst, offset, b.DefaultSampleFlags)
   }
   return offset
}
