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

         // 1. Parse Initialization Segment (still needed for final assembly)
         initData, err := os.ReadFile(initFilePath)
         if err != nil {
            t.Fatalf("Could not read init file: %s, error: %v", initFilePath, err)
         }
         parsedInit, err := ParseFile(initData)
         if err != nil {
            t.Fatalf("Failed to parse init file: %v", err)
         }
         moov := FindMoov(parsedInit)
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
         // 3. Decrypt the segment's mdat boxes in-place.
         keyBytes, err := hex.DecodeString(test.key)
         if err != nil {
            t.Fatalf("Failed to decode test key from hex: %v", err)
         }
         if err := DecryptSegment(parsedSegment, keyBytes); err != nil {
            t.Fatalf("Decryption failed: %v", err)
         }
         // 4. Sanitize metadata and construct the final interleaved MP4
         if err := moov.Sanitize(); err != nil {
            t.Logf("Note: sanitization returned an error (as expected for some clear content): %v", err)
         }
         if trak := moov.GetTrak(); trak != nil {
            trak.RemoveEdts()
         }
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
            finalMP4Data.Write(box.Encode())
         }
         // 5. Write to file and verify
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
