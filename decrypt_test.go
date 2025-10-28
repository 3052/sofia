package sofia

import (
   "bytes"
   "encoding/hex"
   "os"
   "path/filepath"
   "testing"
)

// TestDecryption is a table-driven test that decrypts all provided samples.
func TestDecryption(t *testing.T) {
   const testDataPrefix = "testdata/"
   const outputDir = "test_output"

   if err := os.MkdirAll(outputDir, 0755); err != nil {
      t.Fatalf("Could not create output directory: %v", err)
   }

   for _, test := range senc_tests {
      t.Run(test.out, func(t *testing.T) {
         initFilePath := filepath.Join(testDataPrefix, test.initial)
         segmentFilePath := filepath.Join(testDataPrefix, test.segment)

         // 1. Parse Initialization Segment
         initData, err := os.ReadFile(initFilePath)
         if err != nil {
            t.Fatalf("Could not read init file: %s, error: %v", initFilePath, err)
         }
         parsedInit, err := ParseFile(initData)
         if err != nil {
            t.Fatalf("Failed to parse init file: %v", err)
         }
         var moov *MoovBox
         for i := range parsedInit {
            if parsedInit[i].Moov != nil {
               moov = parsedInit[i].Moov
            }
         }
         if moov == nil {
            t.Fatal("Could not find 'moov' box in init file.")
         }

         // 2. Parse Media Segment
         segmentData, err := os.ReadFile(segmentFilePath)
         if err != nil {
            t.Fatalf("Could not read segment file: %s, error: %v", segmentFilePath, err)
         }
         parsedSegment, err := ParseFile(segmentData)
         if err != nil {
            t.Fatalf("Failed to parse segment file: %v", err)
         }

         // 3. Prepare decryption keys
         var keys KeyMap
         trak := moov.GetTrak()
         if trak == nil {
            t.Fatal("Could not find track in moov box.")
         }
         tenc := trak.GetTenc()
         isEncrypted := tenc != nil

         if isEncrypted {
            keyBytes, err := hex.DecodeString(test.key)
            if err != nil {
               t.Fatalf("Failed to decode test key from hex: %v", err)
            }
            keys = make(KeyMap)
            if err := keys.AddKey(tenc.DefaultKID[:], keyBytes); err != nil {
               t.Fatalf("Failed to add key to KeyMap: %v", err)
            }
         }

         // 4. Decrypt the segment's mdat boxes in-place.
         if err := keys.DecryptSegment(parsedSegment, moov); err != nil {
            t.Fatalf("Decryption failed: %v", err)
         }

         // 5. Sanitize metadata and construct the final interleaved MP4
         if err := moov.Sanitize(); err != nil {
            t.Logf("Note: sanitization returned an error (as expected for some clear content): %v", err)
         }
         trak.RemoveEdts()

         var finalMP4Data bytes.Buffer
         // Write the init segment first
         for _, box := range parsedInit {
            finalMP4Data.Write(box.Encode())
         }

         // Assemble the final file by iterating through the modified segment boxes.
         for _, box := range parsedSegment {
            if box.Moof != nil {
               box.Moof.Sanitize()
            }
            // Encode writes the sanitized moof or the now-decrypted mdat.
            finalMP4Data.Write(box.Encode())
         }

         // 6. Write to file and verify
         outputFilePath := filepath.Join(outputDir, test.out)
         if err := os.WriteFile(outputFilePath, finalMP4Data.Bytes(), 0644); err != nil {
            t.Fatalf("Failed to write final MP4 file: %v", err)
         }

         if bytes.Contains(finalMP4Data.Bytes(), []byte("pssh")) {
            t.Error("'pssh' box found; removal failed.")
         }
         if bytes.Contains(finalMP4Data.Bytes(), []byte("sinf")) {
            t.Error("'sinf' box found; removal failed.")
         }
         if bytes.Contains(finalMP4Data.Bytes(), []byte("edts")) {
            t.Error("'edts' box found; removal failed.")
         }
      })
   }
}
