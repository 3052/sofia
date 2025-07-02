package senc

import (
   "41.neocities.org/sofia"
   "41.neocities.org/sofia/tenc"
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
)

// unknown IV size means the entire sample size is unknown
type Sample struct {
   InitializationVector []uint8
   SubsampleCount       uint16
   Subsample            []Subsample
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func (s *Sample) Decrypt(data, key []byte, tenc_box *tenc.Box) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   var iv [16]byte
   if s.InitializationVector != nil {
      iv = [16]byte(s.InitializationVector)
   } else {
      iv = [16]byte(tenc_box.DefaultConstantIv)
   }
   stream := cipher.NewCTR(block, iv[:])
   if len(s.Subsample) >= 1 {
      var i uint32
      for _, sample1 := range s.Subsample {
         clear := uint32(sample1.BytesOfClearData)
         if clear >= 1 {
            i += clear
         }
         protected := sample1.BytesOfProtectedData
         if protected >= 1 {
            stream.XORKeyStream(data[i:i+protected], data[i:i+protected])
            i += protected
         }
      }
   } else {
      stream.XORKeyStream(data, data)
   }
   return nil
}

// senc_use_subsamples: flag mask is 0x000002.
func (b *Box) senc_use_subsamples() bool {
   return b.FullBoxHeader.GetFlags()&2 >= 1
}

func (s *Sample) Append(data []byte, box1 *Box) ([]byte, error) {
   data = append(data, s.InitializationVector...)
   if box1.senc_use_subsamples() {
      data = binary.BigEndian.AppendUint16(data, s.SubsampleCount)
      for _, sample1 := range s.Subsample {
         var err error
         data, err = sample1.Append(data)
         if err != nil {
            return nil, err
         }
      }
   }
   return data, nil
}
func (s *Sample) Decode(
   data []byte, box1 *Box, tenc_box *tenc.Box,
) (int, error) {
   n := int(tenc_box.DefaultPerSampleIvSize)
   s.InitializationVector = data[:n]
   if box1.senc_use_subsamples() {
      n1, err := binary.Decode(data[n:], binary.BigEndian, &s.SubsampleCount)
      if err != nil {
         return 0, err
      }
      n += n1
      s.Subsample = make([]Subsample, 0, s.SubsampleCount)
      for _, sample1 := range s.Subsample {
         n1, err = sample1.Decode(data[n:])
         if err != nil {
            return 0, err
         }
         n += n1
         s.Subsample = append(s.Subsample, sample1)
      }
   }
   return n, nil
}

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
   Data          []byte
}

func (b *Box) Read(data []byte) error {
   n, err := binary.Decode(data, binary.BigEndian, &b.FullBoxHeader)
   if err != nil {
      return err
   }
   data = data[n:]
   n, err = binary.Decode(data, binary.BigEndian, &b.SampleCount)
   if err != nil {
      return err
   }
   data = data[n:]
   b.Data = data
   return nil
}

func (b *Box) Append(data []byte) ([]byte, error) {
   data, err := b.BoxHeader.Append(data)
   if err != nil {
      return nil, err
   }
   data, err = binary.Append(data, binary.BigEndian, b.FullBoxHeader)
   if err != nil {
      return nil, err
   }
   data = binary.BigEndian.AppendUint32(data, b.SampleCount)
   return append(data, b.Data...), nil
}

///

func (s *Subsample) Append(data []byte) ([]byte, error) {
   return binary.Append(data, binary.BigEndian, s)
}

func (s *Subsample) Decode(data []byte) (int, error) {
   return binary.Decode(data, binary.BigEndian, s)
}

type Subsample struct {
   BytesOfClearData     uint16
   BytesOfProtectedData uint32
}

