package senc

import (
   "41.neocities.org/sofia"
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "log/slog"
)

// ISO/IEC 23001-7
//
// if the version of the SampleEncryptionBox is 0 and the flag
// senc_use_subsamples is set, UseSubSampleEncryption is set to 1
//
//   aligned(8) class SampleEncryptionBox extends FullBox(
//      'senc', version, flags
//   ) {
//      unsigned int(32) sample_count;
//      {
//         unsigned int(Per_Sample_IV_Size*8) InitializationVector;
//         if (UseSubSampleEncryption) {
//            unsigned int(16) subsample_count;
//            {
//               unsigned int(16) BytesOfClearData;
//               unsigned int(32) BytesOfProtectedData;
//            } [subsample_count ]
//         }
//      }[ sample_count ]
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   SampleCount   uint32
   Sample        []Sample
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = b.FullBoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.SampleCount)
   for _, value := range b.Sample {
      data, err = value.Append(data)
      if err != nil {
         return nil, err
      }
   }
   return data, nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (b *Box) senc_use_subsamples() bool {
   return b.FullBoxHeader.GetFlags()&2 >= 1
}

func (s Subsample) Append(data []byte) ([]byte, error) {
   return binary.Append(data, binary.BigEndian, s)
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

func (s *Sample) Append(data []byte) ([]byte, error) {
   data = append(data, s.InitializationVector[:]...)
   if s.box.senc_use_subsamples() {
      data = binary.BigEndian.AppendUint16(data, s.SubsampleCount)
      for _, value := range s.Subsample {
         var err error
         data, err = value.Append(data)
         if err != nil {
            return nil, err
         }
      }
   }
   return data, nil
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (s *Sample) DecryptCenc(text, key []byte) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   var iv [16]byte
   copy(iv[:], s.InitializationVector[:])
   stream := cipher.NewCTR(block, iv[:])
   if len(s.Subsample) >= 1 {
      var i uint32
      for _, sub := range s.Subsample {
         clear := uint32(sub.BytesOfClearData)
         if clear >= 1 {
            i += clear
         }
         protected := sub.BytesOfProtectedData
         if protected >= 1 {
            stream.XORKeyStream(text[i:i+protected], text[i:i+protected])
            i += protected
         }
      }
   } else {
      stream.XORKeyStream(text, text)
   }
   return nil
}

type Sample struct {
   InitializationVector [8]uint8
   SubsampleCount       uint16
   Subsample            []Subsample
   box                  *Box
}

func (b *Box) Read(data []byte) error {
   slog.Debug("senc FullBoxHeader")
   n, err := b.FullBoxHeader.Decode(data)
   if err != nil {
      return err
   }
   data = data[n:]
   slog.Debug("senc SampleCount")
   n, err = binary.Decode(data, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   data = data[n:]
   b.Sample = make([]Sample, b.SampleCount)
   for i, value := range b.Sample {
      value.box = b
      n, err = value.Decode(data)
      if err != nil {
         return err
      }
      data = data[n:]
      b.Sample[i] = value
   }
   return nil
}

func (s *Sample) Decode(data []byte) (int, error) {
   ns := copy(s.InitializationVector[:], data)
   if s.box.senc_use_subsamples() {
      n, err := binary.Decode(data[ns:], binary.BigEndian, &s.SubsampleCount)
      if err != nil {
         return 0, err
      }
      ns += n
      s.Subsample = make([]Subsample, s.SubsampleCount)
      for i, value := range s.Subsample {
         n, err = value.Decode(data[ns:])
         if err != nil {
            return 0, err
         }
         ns += n
         s.Subsample[i] = value
      }
   }
   return ns, nil
}

func (s *Subsample) Decode(data []byte) (int, error) {
   return binary.Decode(data, binary.BigEndian, s)
}
