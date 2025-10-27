package mp4

import (
   "bytes"
   "os"
   "path/filepath"
   "testing"
)

// senc_test defines the structure for our data-driven tests.
type senc_test struct {
   initial string
   key     string
   out     string
   segment string
}

// senc_tests is the table of all files and keys to be used in testing.
var senc_tests = []senc_test{
   {
      initial: "criterion-avc1/0-804.mp4",
      key:     "377772323b0f45efb2c53c603749d834",
      out:     "criterion-avc1.mp4",
      segment: "criterion-avc1/13845-168166.mp4",
   },
   {
      initial: "hboMax-dvh1/0-862.mp4",
      key:     "8ea21645755811ecb84b1f7c39bbbff3",
      out:     "hboMax-dvh1.mp4",
      segment: "hboMax-dvh1/19579-78380.mp4",
   },
   {
      initial: "hboMax-ec-3/0-657.mp4",
      key:     "acaec99945a3615c9ef7b1b04727022a",
      out:     "hboMax-ec-3.mp4",
      segment: "hboMax-ec-3/28710-157870.mp4",
   },
   {
      initial: "hboMax-hvc1/0-834.mp4",
      key:     "a269d5aebc114fd167c380f801437f49",
      out:     "hboMax-hvc1.mp4",
      segment: "hboMax-hvc1/19551-35438.mp4",
   },
   {
      initial: "hulu-avc1/map.mp4",
      key:     "33a7ef13ee16fa6a3d1467c0cc59a84f",
      out:     "hulu-avc1.mp4",
      segment: "hulu-avc1/pts_0.mp4",
   },
   {
      initial: "paramount-mp4a/init.m4v",
      key:     "d98277ff6d7406ec398b49bbd52937d4",
      out:     "paramount-mp4a.mp4",
      segment: "paramount-mp4a/seg_1.m4s",
   },
   {
      initial: "roku-avc1/index_video_8_0_init.mp4",
      key:     "1ba08384626f9523e37b9db17f44da2b",
      out:     "roku-avc1.mp4",
      segment: "roku-avc1/index_video_8_0_1.mp4",
   },
   {
      initial: "rtbf-avc1/vod-idx-3-video=300000.dash",
      key:     "553b091b257584d3938c35dd202531f8",
      out:     "rtbf-avc1.mp4",
      segment: "rtbf-avc1/vod-idx-3-video=300000-0.dash",
   },
   {
      initial: "tubi-avc1/0-1683.mp4",
      key:     "8109222ffe94120d61f887d40d0257ed",
      out:     "tubi-avc1.mp4",
      segment: "tubi-avc1/16524-27006.mp4",
   },
}

// TestRoundTrip is a table-driven test covering all files.
func TestRoundTrip(t *testing.T) {
   const testDataPrefix = "../../testdata/"

   for _, test := range senc_tests {
      t.Run(test.out, func(t *testing.T) {
         filesToTest := []string{test.initial, test.segment}
         for _, file := range filesToTest {
            filePath := filepath.Join(testDataPrefix, file)
            t.Run(filepath.Base(filePath), func(t *testing.T) {
               originalData, err := os.ReadFile(filePath)
               if err != nil {
                  t.Skipf("test file not found, skipping: %s", filePath)
                  return
               }
               if len(originalData) == 0 {
                  return
               }

               parsedBoxes, err := ParseFile(originalData)
               if err != nil {
                  t.Fatalf("ParseFile failed: %v", err)
               }

               var encodedData []byte
               for _, box := range parsedBoxes {
                  encodedData = append(encodedData, box.Encode()...)
               }

               if !bytes.Equal(originalData, encodedData) {
                  t.Errorf("Round trip failed. Original and encoded data do not match.")
               }
            })
         }
      })
   }
}
