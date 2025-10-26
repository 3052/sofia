package mp4

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/hex"
   "errors"
   "fmt"
)

// KeyMap maps a 16-byte Key ID (KID) to its 16-byte decryption key.
type KeyMap map[[16]byte][16]byte

// Decrypter handles the decryption of CENC-encrypted media segments.
type Decrypter struct {
   keys KeyMap
}

// NewDecrypter creates a new decrypter instance.
func NewDecrypter() *Decrypter {
   return &Decrypter{keys: make(KeyMap)}
}

// AddKey adds a decryption key to the decrypter's key map.
func (d *Decrypter) AddKey(kidHex string, keyHex string) error {
   kid, err := hex.DecodeString(kidHex)
   if err != nil || len(kid) != 16 {
      return fmt.Errorf("invalid KID: %w", err)
   }
   key, err := hex.DecodeString(keyHex)
   if err != nil || len(key) != 16 {
      return fmt.Errorf("invalid key: %w", err)
   }
   var kidArray, keyArray [16]byte
   copy(kidArray[:], kid)
   copy(keyArray[:], key)
   d.keys[kidArray] = keyArray
   return nil
}

// Decrypt now correctly checks for the presence of a 'senc' box before attempting decryption.
func (d *Decrypter) Decrypt(moof *MoofBox, mdatData []byte, moov *MoovBox) ([]byte, error) {
   if moof == nil || mdatData == nil || moov == nil {
      return nil, errors.New("moof, mdat, and moov boxes must not be nil")
   }

   decryptedMdat := make([]byte, 0, len(mdatData))
   mdatOffset := 0

   for _, moofChild := range moof.Children {
      traf := moofChild.Traf
      if traf == nil {
         continue
      }

      tfhd := traf.GetTfhd()
      trun := traf.GetTrun()
      senc := traf.GetSenc() // This can be nil for unencrypted segments

      if tfhd == nil || trun == nil {
         return nil, errors.New("traf is missing required boxes: tfhd, trun")
      }

      trak := moov.GetTrakByTrackID(tfhd.TrackID)
      if trak == nil {
         return nil, fmt.Errorf("could not find trak with ID %d", tfhd.TrackID)
      }

      tenc := trak.GetTenc()

      // *** CORRECTED LOGIC ***
      // Only decrypt if both the moov signals encryption (tenc) AND
      // the segment signals encryption (senc).
      if tenc == nil || senc == nil {
         // This segment is in the clear. Copy the data as-is.
         for i, sample := range trun.Samples {
            end := mdatOffset + int(sample.Size)
            if end > len(mdatData) {
               return nil, fmt.Errorf("sample %d size exceeds mdat bounds", i)
            }
            decryptedMdat = append(decryptedMdat, mdatData[mdatOffset:end]...)
            mdatOffset = end
         }
         continue // Move to the next track fragment
      }

      // If we reach here, the segment is definitely encrypted.
      key, ok := d.keys[tenc.DefaultKID]
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
         sampleSize := int(sampleInfo.Size)
         if mdatOffset+sampleSize > len(mdatData) {
            return nil, fmt.Errorf("mdat buffer exhausted at sample %d", i)
         }
         encryptedSample := mdatData[mdatOffset : mdatOffset+sampleSize]
         mdatOffset += sampleSize

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
