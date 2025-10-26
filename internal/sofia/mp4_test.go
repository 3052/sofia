package mp4

import (
   "bytes"
   "os"
   "path/filepath"
   "testing"
)

// TestRoundTrip loops through specified files and ensures that parsing and
// then encoding the data results in the exact same byte sequence.
// This is the clean, final version that uses the module's public API correctly.
func TestRoundTrip(t *testing.T) {
   // The user must place the test files in a 'testdata' subdirectory
   // relative to the module root.
   testFiles := []string{
      "testdata/criterion-avc1/0-804.mp4",
      "testdata/criterion-avc1/13845-168166.mp4",
      "testdata/hboMax-dvh1/0-862.mp4",
      "testdata/hboMax-dvh1/19579-78380.mp4",
      "testdata/hulu-avc1/map.mp4",
      "testdata/hulu-avc1/pts_0.mp4",
      "testdata/paramount-mp4a/init.m4v",
      "testdata/paramount-mp4a/seg_1.m4s",
      "testdata/roku-avc1/index_video_8_0_1.mp4",
      "testdata/tubi-avc1/0-1683.mp4",
      "testdata/tubi-avc1/16524-27006.mp4",
   }

   for _, filePath := range testFiles {
      // Use a subtest for each file for clearer test output.
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

         // 1. PARSE: Use the simple, clean public API.
         // The module now handles all the parsing logic internally.
         parsedBoxes, err := ParseFile(originalData)
         if err != nil {
            t.Fatalf("ParseFile failed for %s: %v", filePath, err)
         }

         // 2. ENCODE: Iterate through the generic Box slice and encode each one.
         // The generic Encode() method handles dispatching to the correct
         // underlying box type.
         var encodedData []byte
         for _, box := range parsedBoxes {
            encodedData = append(encodedData, box.Encode()...)
         }

         // 3. VERIFY: Compare the results.
         if !bytes.Equal(originalData, encodedData) {
            t.Errorf("Round trip failed for %s. Original and encoded data do not match.", filePath)

            // For debugging, you can uncomment these lines to write the output files
            // and inspect the differences with a hex editor.
            // debugDir := "debug_output"
            // os.Mkdir(debugDir, 0755)
            // baseName := filepath.Base(filePath)
            // os.WriteFile(filepath.Join(debugDir, "original_"+baseName), originalData, 0644)
            // os.WriteFile(filepath.Join(debugDir, "encoded_"+baseName), encodedData, 0644)
         }
      })
   }
}
