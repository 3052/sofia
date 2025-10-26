package mp4

import (
   "crypto/aes"
   "crypto/cipher"
   "encoding/hex"
   "errors"
   "fmt"
   "log"
)

// KeyMap, Decrypter, NewDecrypter, AddKey remain the same...
type KeyMap map[[16]byte][16]byte
type Decrypter struct{ keys KeyMap }

func NewDecrypter() *Decrypter { return &Decrypter{keys: make(KeyMap)} }
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

// Decrypt now correctly pads the 8-byte IV into the first half of the 16-byte block.
func (d *Decrypter) Decrypt(moof *MoofBox, mdatData []byte, moov *MoovBox) ([]byte, error) {
   if moof == nil || mdatData == nil || moov == nil {
      return nil, errors.New("moof, mdat, and moov boxes must not be nil")
   }
   log.Printf("[DECRYPT] Starting decryption. MDAT size: %d", len(mdatData))

   decryptedMdat := make([]byte, 0, len(mdatData))
   mdatOffset := 0

   for _, moofChild := range moof.Children {
      traf := moofChild.Traf
      if traf == nil {
         continue
      }

      tfhd, trun, senc := traf.GetTfhd(), traf.GetTrun(), traf.GetSenc()
      if tfhd == nil || trun == nil || senc == nil {
         return nil, errors.New("traf is missing required boxes")
      }
      log.Printf("[DECRYPT] Processing TRAF for TrackID: %d", tfhd.TrackID)

      trak := moov.GetTrakByTrackID(tfhd.TrackID)
      if trak == nil {
         return nil, fmt.Errorf("could not find trak with ID %d", tfhd.TrackID)
      }
      tenc := trak.GetTenc()
      if tenc == nil {
         log.Printf("[DECRYPT] Track %d is not encrypted. Copying data as-is.", tfhd.TrackID)
         for i, sample := range trun.Samples {
            end := mdatOffset + int(sample.Size)
            if end > len(mdatData) {
               return nil, fmt.Errorf("sample %d size exceeds mdat bounds", i)
            }
            decryptedMdat = append(decryptedMdat, mdatData[mdatOffset:end]...)
            mdatOffset = end
         }
         continue
      }

      key, ok := d.keys[tenc.DefaultKID]
      if !ok {
         return nil, fmt.Errorf("no key for KID %s", hex.EncodeToString(tenc.DefaultKID[:]))
      }
      log.Printf("[DECRYPT] Found key for KID %s", hex.EncodeToString(tenc.DefaultKID[:]))

      block, err := aes.NewCipher(key[:])
      if err != nil {
         return nil, fmt.Errorf("AES cipher error: %w", err)
      }
      if len(trun.Samples) != len(senc.Samples) {
         return nil, errors.New("sample count mismatch")
      }

      for i, sampleInfo := range trun.Samples {
         log.Printf("\n--- [DECRYPT] Processing Sample %d ---", i)
         sampleSize := int(sampleInfo.Size)
         log.Printf("  [TRUN] Sample Size: %d", sampleSize)

         if mdatOffset+sampleSize > len(mdatData) {
            return nil, fmt.Errorf("mdat buffer exhausted at sample %d", i)
         }
         encryptedSample := mdatData[mdatOffset : mdatOffset+sampleSize]
         log.Printf("  [MDAT] Current Offset: %d. Reading sample from mdat[%d:%d]", mdatOffset, mdatOffset, mdatOffset+sampleSize)
         mdatOffset += sampleSize

         iv := senc.Samples[i].IV
         log.Printf("  [SENC] IV: %s", hex.EncodeToString(iv))
         if len(iv) == 8 {
            // *** FIX: The 8-byte IV must be in the FIRST half of the 16-byte slice. ***
            paddedIV := make([]byte, 16)
            copy(paddedIV, iv) // Copies into paddedIV[0:8], leaving the rest as 0. This is the fix.
            iv = paddedIV
            log.Printf("  [SENC] Padded IV to 16 bytes: %s", hex.EncodeToString(iv))
         } else if len(iv) != 16 {
            return nil, fmt.Errorf("invalid IV length: %d", len(iv))
         }

         stream := cipher.NewCTR(block, iv)
         decryptedSample := make([]byte, 0, sampleSize)
         sampleOffset := 0

         if len(senc.Samples[i].Subsamples) == 0 {
            log.Printf("    No subsamples. Decrypting all %d bytes.", len(encryptedSample))
            decryptedPortion := make([]byte, len(encryptedSample))
            stream.XORKeyStream(decryptedPortion, encryptedSample)
            decryptedSample = append(decryptedSample, decryptedPortion...)
         } else {
            log.Printf("    Processing %d subsamples (interleaved)...", len(senc.Samples[i].Subsamples))
            for j, sub := range senc.Samples[i].Subsamples {
               clearSize := int(sub.BytesOfClearData)
               protectedSize := int(sub.BytesOfProtectedData)
               log.Printf("    Subsample %d: Clear=%d, Protected=%d. Current sample offset=%d", j, clearSize, protectedSize, sampleOffset)

               endOfClear := sampleOffset + clearSize
               log.Printf("      Copying clear data from sample[%d:%d]", sampleOffset, endOfClear)
               decryptedSample = append(decryptedSample, encryptedSample[sampleOffset:endOfClear]...)
               sampleOffset = endOfClear

               if protectedSize > 0 {
                  endOfProtected := sampleOffset + protectedSize
                  log.Printf("      Decrypting protected data from sample[%d:%d]", sampleOffset, endOfProtected)
                  protectedData := encryptedSample[sampleOffset:endOfProtected]
                  stream.XORKeyStream(protectedData, protectedData)
                  decryptedSample = append(decryptedSample, protectedData...)
                  sampleOffset = endOfProtected
               } else {
                  log.Printf("      Skipping decryption for subsample with 0 protected bytes.")
               }
            }
         }
         decryptedMdat = append(decryptedMdat, decryptedSample...)
         log.Printf("  [MDAT] New Offset: %d", mdatOffset)
      }
   }
   return decryptedMdat, nil
}
