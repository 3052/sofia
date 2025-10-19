// File: trun_box.go
package mp4parser

type TrunBox struct {
   FullBox
   SampleCount      uint32
   DataOffset       int32
   FirstSampleFlags uint32
   Samples          []TrunSample
}
type TrunSample struct {
   SampleDuration              uint32
   SampleSize                  uint32
   SampleFlags                 uint32
   SampleCompositionTimeOffset int32
}

func ParseTrunBox(data []byte) (*TrunBox, error) {
   b := &TrunBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   b.SampleCount, offset, err = readUint32(data, offset)
   if err != nil {
      return nil, err
   }
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      var val uint32
      val, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
      b.DataOffset = int32(val)
   }
   if flags&0x000004 != 0 {
      b.FirstSampleFlags, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
   }
   b.Samples = make([]TrunSample, b.SampleCount)
   for i := 0; i < int(b.SampleCount); i++ {
      sample := TrunSample{}
      if flags&0x000100 != 0 {
         sample.SampleDuration, offset, err = readUint32(data, offset)
         if err != nil {
            return nil, err
         }
      }
      if flags&0x000200 != 0 {
         sample.SampleSize, offset, err = readUint32(data, offset)
         if err != nil {
            return nil, err
         }
      }
      if flags&0x000400 != 0 {
         sample.SampleFlags, offset, err = readUint32(data, offset)
         if err != nil {
            return nil, err
         }
      }
      if flags&0x000800 != 0 {
         var val uint32
         val, offset, err = readUint32(data, offset)
         if err != nil {
            return nil, err
         }
         sample.SampleCompositionTimeOffset = int32(val)
      }
      b.Samples[i] = sample
   }
   return b, nil
}
func (b *TrunBox) Size() uint64 {
   size := uint64(8) + b.FullBox.Size() + 4
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      size += 4
   }
   if flags&0x000004 != 0 {
      size += 4
   }
   sampleSize := uint64(0)
   if flags&0x000100 != 0 {
      sampleSize += 4
   }
   if flags&0x000200 != 0 {
      sampleSize += 4
   }
   if flags&0x000400 != 0 {
      sampleSize += 4
   }
   if flags&0x000800 != 0 {
      sampleSize += 4
   }
   size += uint64(len(b.Samples)) * sampleSize
   return size
}
func (b *TrunBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "trun")
   offset = b.FullBox.Format(dst, offset)
   offset = writeUint32(dst, offset, b.SampleCount)
   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   if flags&0x000001 != 0 {
      offset = writeUint32(dst, offset, uint32(b.DataOffset))
   }
   if flags&0x000004 != 0 {
      offset = writeUint32(dst, offset, b.FirstSampleFlags)
   }
   for _, sample := range b.Samples {
      if flags&0x000100 != 0 {
         offset = writeUint32(dst, offset, sample.SampleDuration)
      }
      if flags&0x000200 != 0 {
         offset = writeUint32(dst, offset, sample.SampleSize)
      }
      if flags&0x000400 != 0 {
         offset = writeUint32(dst, offset, sample.SampleFlags)
      }
      if flags&0x000800 != 0 {
         offset = writeUint32(dst, offset, uint32(sample.SampleCompositionTimeOffset))
      }
   }
   return offset
}
