package sofia

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/binary"
   "io"
)

// 8.1.1 Media data box
//  aligned(8) class MediaDataBox extends Box('mdat') {
//     bit(8) data[];
//  }
type MediaDataBox struct {
   Header BoxHeader
   Data   [][]byte
}

func (b *MediaDataBox) Decode(t TrackRunBox, r io.Reader) error {
   b.Data = make([][]byte, t.Sample_Count)
   for i := range b.Data {
      data := make([]byte, t.Samples[i].Size)
      _, err := io.ReadFull(r, data)
      if err != nil {
         return err
      }
      b.Data[i] = data
   }
   return nil
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func CryptSampleCenc(sample, key []byte, enc EncryptionSample) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   var iv [16]byte
   binary.BigEndian.PutUint64(iv[:], enc.InitializationVector)
   stream := cipher.NewCTR(block, iv[:])
   if len(enc.Subsamples) >= 1 {
      var pos uint32
      for _, ss := range enc.Subsamples {
         nrClear := uint32(ss.BytesOfClearData)
         if nrClear >= 1 {
            pos += nrClear
         }
         nrEnc := ss.BytesOfProtectedData
         if nrEnc >= 1 {
            stream.XORKeyStream(sample[pos:pos+nrEnc], sample[pos:pos+nrEnc])
            pos += nrEnc
         }
      }
   } else {
      stream.XORKeyStream(sample, sample)
   }
   return nil
}

func (b MediaDataBox) Encode(w io.Writer) error {
   err := b.Header.Encode(w)
   if err != nil {
      return err
   }
   for _, data := range b.Data {
      _, err := w.Write(data)
      if err != nil {
         return err
      }
   }
   return nil
}
