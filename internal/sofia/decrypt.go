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
   return &Decrypter{
      keys: make(KeyMap),
   }
}

// AddKey adds a decryption key to the decrypter's key map.
func (d *Decrypter) AddKey(kidHex string, keyHex string) error {
   kid, err := hex.DecodeString(kidHex)
   if err != nil || len(kid) != 16 {
      return fmt.Errorf("invalid KID hex string: %w", err)
   }
   key, err := hex.DecodeString(keyHex)
   if err != nil || len(key) != 16 {
      return fmt.Errorf("invalid key hex string: %w", err)
   }

   var kidArray [16]byte
   var keyArray [16]byte
   copy(kidArray[:], kid)
   copy(keyArray[:], key)

   d.keys[kidArray] = keyArray
   return nil
}

// Decrypt takes a parsed movie fragment (moof) and its corresponding media data (mdat),
// along with the movie's initialization data (moov), and returns the decrypted media data.
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
      senc := traf.GetSenc()
      if tfhd == nil || trun == nil || senc == nil {
         return nil, errors.New("traf is missing one or more required boxes: tfhd, trun, senc")
      }

      trak := moov.GetTrakByTrackID(tfhd.TrackID)
      if trak == nil {
         return nil, fmt.Errorf("could not find trak with ID %d in moov", tfhd.TrackID)
      }
      tenc := trak.GetTenc()
      if tenc == nil {
         for _, sample := range trun.Samples {
            end := mdatOffset + int(sample.Size)
            if end > len(mdatData) {
               return nil, fmt.Errorf("sample size exceeds mdat bounds")
            }
            decryptedMdat = append(decryptedMdat, mdatData[mdatOffset:end]...)
            mdatOffset = end
         }
         continue
      }

      key, ok := d.keys[tenc.DefaultKID]
      if !ok {
         return nil, fmt.Errorf("no key found for KID %s", hex.EncodeToString(tenc.DefaultKID[:]))
      }

      block, err := aes.NewCipher(key[:])
      if err != nil {
         return nil, fmt.Errorf("could not create AES cipher: %w", err)
      }

      if len(trun.Samples) != len(senc.Samples) {
         return nil, errors.New("sample count mismatch between trun and senc boxes")
      }

      for i, sampleInfo := range trun.Samples {
         sampleSize := int(sampleInfo.Size)
         if mdatOffset+sampleSize > len(mdatData) {
            return nil, errors.New("mdat buffer exhausted; sample size larger than remaining data")
         }
         encryptedSample := mdatData[mdatOffset : mdatOffset+sampleSize]
         mdatOffset += sampleSize

         // *** FIX: Handle 8-byte IVs by padding them to 16 bytes ***
         iv := senc.Samples[i].IV
         if len(iv) == 8 {
            // Pad 8-byte IV to 16 bytes by prepending zeroes
            paddedIV := make([]byte, 16)
            copy(paddedIV[8:], iv)
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
               decryptedSample = append(decryptedSample, encryptedSample[sampleOffset:sampleOffset+sub.BytesOfClearData]...)
               sampleOffset += sub.BytesOfClearData

               protectedData := encryptedSample[sampleOffset : sampleOffset+sub.BytesOfProtectedData]
               decryptedPortion := make([]byte, len(protectedData))
               stream.XORKeyStream(decryptedPortion, protectedData)
               decryptedSample = append(decryptedSample, decryptedPortion...)
               sampleOffset += sub.BytesOfProtectedData
            }
         }
         decryptedMdat = append(decryptedMdat, decryptedSample...)
      }
   }
   return decryptedMdat, nil
}
