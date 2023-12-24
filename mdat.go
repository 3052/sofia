package sofia

import (
   "crypto/aes"
   "crypto/cipher"
   "io"
)

// aligned(8) class MediaDataBox extends Box('mdat') {
//    bit(8) data[];
// }
type MediaDataBox struct {
   Header BoxHeader
   Data   [][]byte
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

func (b *MediaDataBox) Decode(t TrackRunBox, r io.Reader) error {
   for _, sample := range t.Samples {
      data := make([]byte, sample.Size)
      _, err := r.Read(data)
      if err != nil {
         return err
      }
      b.Data = append(b.Data, data)
   }
   return nil
}

// github.com/Eyevinn/mp4ff/blob/v0.40.2/mp4/crypto.go#L101
func CryptSampleCenc(sample, key []byte, enc EncryptionSample) error {
   block, err := aes.NewCipher(key)
   if err != nil {
      return err
   }
   stream := cipher.NewCTR(block, enc.InitializationVector[:])
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

