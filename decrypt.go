package sofia

import (
   "crypto/aes"
   "crypto/cipher"
   "errors"
   "fmt"
)

// DecryptSegment decrypts a segment's mdat boxes in-place using the provided key.
// It self-determines if decryption is needed by checking for 'senc' boxes within the segment.
// This function has a side effect: it modifies the Payload of the MdatBox structs within the segmentBoxes slice.
func DecryptSegment(segmentBoxes []Box, key []byte) error {
   // First, check if any part of this segment is actually encrypted.
   var isEncrypted bool
   for _, box := range segmentBoxes {
      if box.Moof != nil {
         for _, child := range box.Moof.Children {
            if child.Traf != nil {
               if _, ok := child.Traf.GetSenc(); ok {
                  isEncrypted = true
                  break
               }
            }
         }
      }
      if isEncrypted {
         break
      }
   }

   // If no 'senc' boxes were found in any fragment, there is nothing to decrypt.
   if !isEncrypted {
      return nil
   }

   // If the segment is encrypted, we must have a valid key.
   if len(key) != 16 {
      return fmt.Errorf("invalid key length: expected 16, got %d", len(key))
   }

   block, err := aes.NewCipher(key)
   if err != nil {
      return fmt.Errorf("AES cipher error: %w", err)
   }

   // Iterate through the boxes, processing moof/mdat pairs.
   for i := 0; i < len(segmentBoxes); i++ {
      if segmentBoxes[i].Moof != nil {
         moof := segmentBoxes[i].Moof
         if i+1 >= len(segmentBoxes) || segmentBoxes[i+1].Mdat == nil {
            return fmt.Errorf("malformed segment: moof at index %d is not followed by an mdat", i)
         }
         mdat := segmentBoxes[i+1].Mdat

         // Perform decryption directly on the MdatBox's payload.
         err := decryptFragment(moof, mdat.Payload, block)
         if err != nil {
            return fmt.Errorf("failed to process fragment at index %d: %w", i, err)
         }
         i++ // Skip the mdat box in the next iteration.
      }
   }
   return nil
}

// decryptFragment is an unexported helper that handles a single moof/mdat pair, decrypting the mdatData in-place.
func decryptFragment(moof *MoofBox, mdatData []byte, block cipher.Block) error {
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
      // The 'senc' box is the per-fragment signal. If it's not here, we skip this traf.
      senc, ok := traf.GetSenc()
      if !ok {
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
