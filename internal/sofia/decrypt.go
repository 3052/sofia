package mp4

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/hex"
   "errors"
   "fmt"
)

// KeyMap now holds the keys and all associated methods.
type KeyMap map[[16]byte][16]byte

// AddKey adds a decryption key to the key map.
func (km KeyMap) AddKey(kid []byte, key []byte) error {
   if len(kid) != 16 {
      return fmt.Errorf("invalid KID length: expected 16, got %d", len(kid))
   }
   if len(key) != 16 {
      return fmt.Errorf("invalid key length: expected 16, got %d", len(key))
   }

   var kidArray [16]byte
   var keyArray [16]byte
   copy(kidArray[:], kid)
   copy(keyArray[:], key)

   km[kidArray] = keyArray
   return nil
}

// Decrypt is now a method on KeyMap, using the keys it contains.
func (km KeyMap) Decrypt(moof *MoofBox, mdatData []byte, moov *MoovBox) ([]byte, error) {
   if moof == nil || mdatData == nil || moov == nil {
      return nil, errors.New("moof, mdat, and moov boxes must not be nil")
   }
   if km == nil {
      return nil, errors.New("keyMap cannot be nil")
   }

   decryptedMdat := make([]byte, 0, len(mdatData))
   mdatOffset := 0

   for _, moofChild := range moof.Children {
      traf := moofChild.Traf
      if traf == nil {
         continue
      }

      tfhd, trun, senc := traf.GetTfhd(), traf.GetTrun(), traf.GetSenc()
      if tfhd == nil || trun == nil {
         return nil, errors.New("traf is missing required boxes: tfhd, trun")
      }

      trak := moov.GetTrakByTrackID(tfhd.TrackID)
      if trak == nil {
         return nil, fmt.Errorf("could not find trak with ID %d", tfhd.TrackID)
      }

      tenc := trak.GetTenc()

      if tenc == nil || senc == nil {
         for i, sample := range trun.Samples {
            sampleSize := sample.Size
            if sampleSize == 0 {
               sampleSize = tfhd.DefaultSampleSize
            }
            if sampleSize == 0 {
               return nil, fmt.Errorf("sample %d has zero size and no default is available", i)
            }
            end := mdatOffset + int(sampleSize)
            if end > len(mdatData) {
               return nil, fmt.Errorf("sample %d size exceeds mdat bounds", i)
            }
            decryptedMdat = append(decryptedMdat, mdatData[mdatOffset:end]...)
            mdatOffset = end
         }
         continue
      }

      key, ok := km[tenc.DefaultKID]
      if !ok {
         return nil, fmt.Errorf("no key for KID %s", hex.EncodeToString(tenc.DefaultKID[:]))
      }

      block, err := aes.NewCipher(key[:])
      if err != nil {
         return nil, fmt.Errorf("AES cipher error: %w", err)
      }

      if len(trun.Samples) != len(senc.Samples) {
         return nil, errors.New("sample count mismatch between trun and senc")
      }

      for i, sampleInfo := range trun.Samples {
         sampleSize := sampleInfo.Size
         if sampleSize == 0 {
            sampleSize = tfhd.DefaultSampleSize
         }
         if sampleSize == 0 {
            return nil, fmt.Errorf("encrypted sample %d has zero size and no default is available", i)
         }
         if mdatOffset+int(sampleSize) > len(mdatData) {
            return nil, fmt.Errorf("mdat buffer exhausted at sample %d", i)
         }
         encryptedSample := mdatData[mdatOffset : mdatOffset+int(sampleSize)]
         mdatOffset += int(sampleSize)

         iv := senc.Samples[i].IV
         if len(iv) == 8 {
            paddedIV := make([]byte, 16)
            copy(paddedIV, iv)
            iv = paddedIV
         } else if len(iv) != 16 {
            return nil, fmt.Errorf("invalid IV length: got %d, want 16", len(iv))
         }

         stream := cipher.NewCTR(block, iv)
         decryptedSample := make([]byte, 0, sampleSize)
         sampleOffset := 0

         if len(senc.Samples[i].Subsamples) == 0 {
            decryptedPortion := make([]byte, len(encryptedSample))
            stream.XORKeyStream(decryptedPortion, encryptedSample)
            decryptedSample = append(decryptedSample, decryptedPortion...)
         } else {
            for _, sub := range senc.Samples[i].Subsamples {
               clearSize := int(sub.BytesOfClearData)
               protectedSize := int(sub.BytesOfProtectedData)
               endOfClear := sampleOffset + clearSize
               decryptedSample = append(decryptedSample, encryptedSample[sampleOffset:endOfClear]...)
               sampleOffset = endOfClear
               if protectedSize > 0 {
                  endOfProtected := sampleOffset + protectedSize
                  protectedData := encryptedSample[sampleOffset:endOfProtected]
                  stream.XORKeyStream(protectedData, protectedData)
                  decryptedSample = append(decryptedSample, protectedData...)
                  sampleOffset = endOfProtected
               }
            }
         }
         decryptedMdat = append(decryptedMdat, decryptedSample...)
      }
   }
   return decryptedMdat, nil
}
