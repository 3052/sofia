package sofia

import (
   "crypto/aes"
   "crypto/cipher"
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

// DecryptSegment processes a slice of parsed boxes, decrypting the mdat payloads in-place.
// This function has a side effect: it modifies the Payload of the MdatBox structs within the segmentBoxes slice.
func (km KeyMap) DecryptSegment(segmentBoxes []Box, moov *MoovBox) error {
   if moov == nil {
      return errors.New("moov box must not be nil")
   }

   trak := moov.GetTrak()
   if trak == nil {
      return errors.New("could not find trak in moov")
   }

   tenc := trak.GetTenc()
   if tenc == nil {
      // Content is not encrypted, nothing to do.
      return nil
   }

   key, ok := km[tenc.DefaultKID]
   if !ok {
      return fmt.Errorf("no key for KID %x", tenc.DefaultKID)
   }

   block, err := aes.NewCipher(key[:])
   if err != nil {
      return fmt.Errorf("AES cipher error: %w", err)
   }

   // Iterate through the boxes, processing moof/mdat pairs as they are found.
   for i := 0; i < len(segmentBoxes); i++ {
      if segmentBoxes[i].Moof != nil {
         moof := segmentBoxes[i].Moof
         if i+1 >= len(segmentBoxes) || segmentBoxes[i+1].Mdat == nil {
            return fmt.Errorf("malformed segment: moof at index %d is not followed by an mdat", i)
         }
         mdat := segmentBoxes[i+1].Mdat

         // Perform the decryption directly on the MdatBox's payload.
         err := km.decryptFragment(moof, mdat.Payload, block)
         if err != nil {
            return fmt.Errorf("failed to process fragment at index %d: %w", i, err)
         }
         i++ // Skip the mdat box in the next iteration.
      }
   }
   return nil
}

// decryptFragment handles a single moof/mdat pair, decrypting the mdatData in-place.
func (km KeyMap) decryptFragment(moof *MoofBox, mdatData []byte, block cipher.Block) error {
   currentMdatOffset := 0

   for _, moofChild := range moof.Children {
      traf := moofChild.Traf
      if traf == nil {
         continue
      }

      tfhd, trun := traf.GetTfhd(), traf.GetTrun()
      if tfhd == nil || trun == nil {
         return errors.New("traf is missing required boxes: tfhd or trun")
      }
      senc := traf.GetSenc()
      if senc == nil {
         // This fragment is not encrypted, so we do nothing.
         continue
      }

      if len(trun.Samples) != len(senc.Samples) {
         return errors.New("sample count mismatch between trun and senc")
      }

      for i, sampleInfo := range trun.Samples {
         sampleSize := sampleInfo.Size
         if sampleSize == 0 {
            sampleSize = tfhd.DefaultSampleSize
         }
         if sampleSize == 0 {
            return fmt.Errorf("sample %d has zero size", i)
         }
         if currentMdatOffset+int(sampleSize) > len(mdatData) {
            return fmt.Errorf("mdat buffer exhausted at sample %d", i)
         }
         sampleData := mdatData[currentMdatOffset : currentMdatOffset+int(sampleSize)]
         currentMdatOffset += int(sampleSize)

         iv := senc.Samples[i].IV
         if len(iv) == 8 {
            paddedIV := make([]byte, 16)
            copy(paddedIV, iv)
            iv = paddedIV
         } else if len(iv) != 16 {
            return fmt.Errorf("invalid IV length for sample %d: got %d, want 16", i, len(iv))
         }

         stream := cipher.NewCTR(block, iv)

         if len(senc.Samples[i].Subsamples) == 0 {
            // Full sample encryption, decrypt in-place.
            stream.XORKeyStream(sampleData, sampleData)
         } else {
            // Subsample encryption.
            sampleOffset := 0
            for _, sub := range senc.Samples[i].Subsamples {
               sampleOffset += int(sub.BytesOfClearData)
               if sub.BytesOfProtectedData > 0 {
                  endOfProtected := sampleOffset + int(sub.BytesOfProtectedData)
                  protectedPortion := sampleData[sampleOffset:endOfProtected]
                  // Decrypt the protected portion in-place.
                  stream.XORKeyStream(protectedPortion, protectedPortion)
                  sampleOffset = endOfProtected
               }
            }
         }
      }
   }
   return nil
}
