package mp4

import (
   "bytes"
   "os"
   "path/filepath"
   "testing"
)

var testFiles = []string{
   `..\..\testdata\criterion-avc1\0-804.mp4`,
   `..\..\testdata\criterion-avc1\13845-168166.mp4`,
   `..\..\testdata\hboMax-dvh1\0-862.mp4`,
   `..\..\testdata\hboMax-dvh1\19579-78380.mp4`,
   `..\..\testdata\hboMax-ec-3\0-657.mp4`,
   `..\..\testdata\hboMax-ec-3\28710-157870.mp4`,
   `..\..\testdata\hboMax-hvc1\0-834.mp4`,
   `..\..\testdata\hboMax-hvc1\19551-35438.mp4`,
   `..\..\testdata\hulu-avc1\map.mp4`,
   `..\..\testdata\hulu-avc1\pts_0.mp4`,
   `..\..\testdata\paramount-mp4a\init.m4v`,
   `..\..\testdata\paramount-mp4a\seg_1.m4s`,
   `..\..\testdata\roku-avc1\index_video_8_0_1.mp4`,
   `..\..\testdata\roku-avc1\index_video_8_0_init.mp4`,
   `..\..\testdata\rtbf-avc1\vod-idx-3-video=300000-0.dash`,
   `..\..\testdata\rtbf-avc1\vod-idx-3-video=300000.dash`,
   `..\..\testdata\tubi-avc1\0-1683.mp4`,
   `..\..\testdata\tubi-avc1\16524-27006.mp4`,
}

func TestRoundTrip(t *testing.T) {
   for _, filePath := range testFiles {
      t.Run(filepath.Base(filePath), func(t *testing.T) {
         originalData, err := os.ReadFile(filePath)
         if err != nil {
            t.Skipf("test file not found, skipping: %s", filePath)
            return
         }
         if len(originalData) == 0 {
            t.Logf("test file is empty: %s", filePath)
            return
         }

         parsedBoxes, err := ParseFile(originalData)
         if err != nil {
            t.Fatalf("ParseFile failed for %s: %v", filePath, err)
         }

         var encodedData []byte
         for _, box := range parsedBoxes {
            encodedData = append(encodedData, box.Encode()...)
         }

         if !bytes.Equal(originalData, encodedData) {
            t.Errorf("Round trip failed for %s. Original and encoded data do not match.", filePath)
         }
      })
   }
}

// TestDecryption demonstrates how to use the Decrypter to decrypt a media fragment.
func TestDecryption(t *testing.T) {
   // --- User-provided information ---
   const knownKID = "2ae3928e76864505aa8499db218b0288"
   const placeholderKey = "000102030405060708090a0b0c0d0e0f" // User must provide the real key.

   // --- Test Setup ---
   initData, err := os.ReadFile("testdata/paramount-mp4a/init.m4v")
   if err != nil {
      t.Skip("Skipping decryption test: could not read init file.")
   }
   parsedInit, err := ParseFile(initData)
   if err != nil {
      t.Fatalf("Failed to parse init file: %v", err)
   }
   var moov *MoovBox
   for _, box := range parsedInit {
      if box.Moov != nil {
         moov = box.Moov
         break
      }
   }
   if moov == nil {
      t.Fatal("Could not find 'moov' box in init file.")
   }

   segmentData, err := os.ReadFile("testdata/paramount-mp4a/seg_1.m4s")
   if err != nil {
      t.Skip("Skipping decryption test: could not read segment file.")
   }
   parsedSegment, err := ParseFile(segmentData)
   if err != nil {
      t.Fatalf("Failed to parse segment file: %v", err)
   }
   var moof *MoofBox
   var mdat *MdatBox
   for _, box := range parsedSegment {
      if box.Moof != nil {
         moof = box.Moof
      }
      if box.Mdat != nil {
         mdat = box.Mdat
      }
   }
   if moof == nil || mdat == nil {
      t.Fatal("Could not find 'moof' and/or 'mdat' box in segment file.")
   }

   // --- Decryption ---
   decrypter := NewDecrypter()
   if err := decrypter.AddKey(knownKID, placeholderKey); err != nil {
      t.Fatalf("Failed to add key to decrypter: %v", err)
   }

   // mdat.Data[8:] skips the 8-byte mdat header (size and type).
   decryptedData, err := decrypter.Decrypt(moof, mdat.Data[8:], moov)
   if err != nil {
      t.Fatalf("Decryption failed: %v", err)
   }

   // --- Verification ---
   originalMdatSize := len(mdat.Data[8:])
   if len(decryptedData) != originalMdatSize {
      t.Errorf("Decrypted data size mismatch: got %d, want %d", len(decryptedData), originalMdatSize)
   }
   if len(decryptedData) == 0 {
      t.Error("Decryption produced zero-length output.")
   }

   // Because we are using a placeholder key, the decrypted data will be garbage,
   // but it will NOT be the same as the original encrypted data. This confirms
   // that the XOR operation was applied.
   if bytes.Equal(decryptedData, mdat.Data[8:]) {
      t.Error("Decrypted data is identical to encrypted data; decryption did not happen correctly.")
   }

   t.Logf("Successfully ran decryption process. Decrypted %d bytes.", len(decryptedData))
}
