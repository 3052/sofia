package mp4parser

// ParseSencContent parses the content of a 'senc' box.
// This function is designed to be called separately from the main parser,
// as it requires the `perSampleIVSize` from the 'tenc' box as context.
func ParseSencContent(data []byte, perSampleIVSize uint8) (*SencBox, error) {
   b := &SencBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   b.SampleCount, offset, err = readUint32(data, offset)
   if err != nil {
      return nil, err
   }

   flags := uint32(b.Flags[0])<<16 | uint32(b.Flags[1])<<8 | uint32(b.Flags[2])
   hasSubsamples := (flags & 0x000002) != 0

   b.InitializationVectors = make([]InitializationVector, b.SampleCount)
   for i := 0; i < int(b.SampleCount); i++ {
      iv := InitializationVector{}

      // Use the provided perSampleIVSize. This is the critical change.
      ivSize := int(perSampleIVSize)
      if ivSize > 0 {
         if offset+ivSize > len(data) {
            return nil, ErrUnexpectedEOF
         }
         iv.IV = data[offset : offset+ivSize]
         offset += ivSize
      }

      if hasSubsamples {
         var subsampleCount uint16
         subsampleCount, offset, err = readUint16(data, offset)
         if err != nil {
            return nil, err
         }
         iv.Subsamples = make([]Subsample, subsampleCount)
         for j := 0; j < int(subsampleCount); j++ {
            var clearData uint16
            clearData, offset, err = readUint16(data, offset)
            if err != nil {
               return nil, err
            }
            var protectedData uint32
            protectedData, offset, err = readUint32(data, offset)
            if err != nil {
               return nil, err
            }
            iv.Subsamples[j] = Subsample{
               BytesOfClearData:     clearData,
               BytesOfProtectedData: protectedData,
            }
         }
      }
      b.InitializationVectors[i] = iv
   }
   return b, nil
}

// SencBox (Sample Encryption Box) holds the parsed sample encryption data.
type SencBox struct {
   FullBox
   SampleCount           uint32
   InitializationVectors []InitializationVector
}

// InitializationVector represents the IV and subsample encryption data for a sample.
type InitializationVector struct {
   IV         []byte
   Subsamples []Subsample
}

// Subsample defines clear and encrypted byte counts within a sample.
type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}
