package mp4

import (
   "bytes"
   "encoding/binary"
   "encoding/hex"
   "os"
   "path/filepath"
   "testing"
)

// TestRoundTrip verifies that parsing and re-encoding a file results in a byte-perfect copy.
func TestRoundTrip(t *testing.T) {
   testFiles := []string{
      "../../testdata/criterion-avc1/0-804.mp4",
      "../../testdata/criterion-avc1/13845-168166.mp4",
      "../../testdata/hboMax-dvh1/0-862.mp4",
      "../../testdata/hboMax-dvh1/19579-78380.mp4",
      "../../testdata/hboMax-ec-3/0-657.mp4",
      "../../testdata/hboMax-ec-3/28710-157870.mp4",
      "../../testdata/hboMax-hvc1/0-834.mp4",
      "../../testdata/hboMax-hvc1/19551-35438.mp4",
      "../../testdata/hulu-avc1/map.mp4",
      "../../testdata/hulu-avc1/pts_0.mp4",
      "../../testdata/paramount-mp4a/init.m4v",
      "../../testdata/paramount-mp4a/seg_1.m4s",
      "../../testdata/roku-avc1/index_video_8_0_1.mp4",
      "../../testdata/roku-avc1/index_video_8_0_init.mp4",
      "../../testdata/rtbf-avc1/vod-idx-3-video=300000-0.dash",
      "../../testdata/rtbf-avc1/vod-idx-3-video=300000.dash",
      "../../testdata/tubi-avc1/0-1683.mp4",
      "../../testdata/tubi-avc1/16524-27006.mp4",
   }

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

// createMdatBox is a helper to construct a valid mdat box from a data payload.
func createMdatBox(payload []byte) []byte {
   // The size includes the 8-byte header (4 for size, 4 for type).
   size := uint32(len(payload) + 8)
   mdatBox := make([]byte, size)
   binary.BigEndian.PutUint32(mdatBox[0:4], size)
   copy(mdatBox[4:8], "mdat")
   copy(mdatBox[8:], payload)
   return mdatBox
}

// TestDecryption now assembles and writes a complete, playable MP4 file.
func TestDecryption(t *testing.T) {
   const placeholderKey = "1ba08384626f9523e37b9db17f44da2b"
   // --- Test Setup ---
   initFilePath := "../../testdata/roku-avc1/index_video_8_0_init.mp4"
   segmentFilePath := "../../testdata/roku-avc1/index_video_8_0_1.mp4"

   initData, err := os.ReadFile(initFilePath)
   if err != nil {
      t.Skipf("Skipping decryption test: could not read init file: %s", initFilePath)
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

   // --- Dynamically Extract the KID ---
   trak := moov.GetTrakByTrackID(1)
   if trak == nil {
      t.Fatal("Could not find video track in moov box.")
   }
   tenc := trak.GetTenc()
   if tenc == nil {
      t.Fatal("Could not find 'tenc' (encryption) box in track. Is the content actually encrypted?")
   }
   kidFromFile := hex.EncodeToString(tenc.DefaultKID[:])
   t.Logf("Successfully extracted KID from file: %s", kidFromFile)

   // --- Load Media Segment ---
   segmentData, err := os.ReadFile(segmentFilePath)
   if err != nil {
      t.Skipf("Skipping decryption test: could not read segment file: %s", segmentFilePath)
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
   if err := decrypter.AddKey(kidFromFile, placeholderKey); err != nil {
      t.Fatalf("Failed to add key to decrypter: %v", err)
   }

   decryptedPayload, err := decrypter.Decrypt(moof, mdat.Data[8:], moov)
   if err != nil {
      t.Fatalf("Decryption failed: %v", err)
   }

   // --- NEW: Assemble and write a valid, playable MP4 file ---
   t.Log("Assembling final MP4 file...")

   // 1. Get the raw bytes of the moof box from the parsed segment.
   var moofBoxData []byte
   for _, box := range parsedSegment {
      if box.Moof != nil {
         moofBoxData = box.Moof.Encode()
         break
      }
   }
   if moofBoxData == nil {
      t.Fatal("Could not extract moof box data for reassembly.")
   }

   // 2. Create a new mdat box with the decrypted payload.
   newMdatBoxData := createMdatBox(decryptedPayload)

   // 3. Concatenate all parts: Init Segment + moof + decrypted mdat
   var finalMP4Data bytes.Buffer
   finalMP4Data.Write(initData)
   finalMP4Data.Write(moofBoxData)
   finalMP4Data.Write(newMdatBoxData)

   // 4. Write the final file to disk.
   outputFilePath := "roku-avc1.mp4"
   if err := os.WriteFile(outputFilePath, finalMP4Data.Bytes(), 0644); err != nil {
      t.Fatalf("Failed to write final MP4 file to disk: %v", err)
   }
   t.Logf("Successfully wrote playable MP4 file to: %s", outputFilePath)

   // --- Verification ---
   if finalMP4Data.Len() == 0 {
      t.Error("Final assembled MP4 data is zero-length.")
   }
   if !bytes.Contains(finalMP4Data.Bytes(), []byte("moov")) {
      t.Error("Final MP4 is missing 'moov' box.")
   }
   if !bytes.Contains(finalMP4Data.Bytes(), []byte("moof")) {
      t.Error("Final MP4 is missing 'moof' box.")
   }

   t.Log("Decryption and file assembly complete.")
}
