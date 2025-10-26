package mp4

import (
   "bytes"
   "os"
   "path/filepath"
   "testing"
)

var testFiles = []string{
   `criterion-avc1\0-804.mp4`,
   `criterion-avc1\13845-168166.mp4`,
   `hboMax-dvh1\0-862.mp4`,
   `hboMax-dvh1\19579-78380.mp4`,
   `hboMax-ec-3\0-657.mp4`,
   `hboMax-ec-3\28710-157870.mp4`,
   `hboMax-hvc1\0-834.mp4`,
   `hboMax-hvc1\19551-35438.mp4`,
   `hulu-avc1\map.mp4`,
   `hulu-avc1\pts_0.mp4`,
   `paramount-mp4a\init.m4v`,
   `paramount-mp4a\seg_1.m4s`,
   `roku-avc1\index_video_8_0_1.mp4`,
   `roku-avc1\index_video_8_0_init.mp4`,
   `rtbf-avc1\vod-idx-3-video=300000-0.dash`,
   `rtbf-avc1\vod-idx-3-video=300000.dash`,
   `tubi-avc1\0-1683.mp4`,
   `tubi-avc1\16524-27006.mp4`,
}

const folder = "../../testdata/"

func TestRoundTrip(t *testing.T) {
   for _, filePath := range testFiles {
      // Use a subtest for each file for clearer test output.
      t.Run(filepath.Base(filePath), func(t *testing.T) {
         originalData, err := os.ReadFile(folder + filePath)
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
