package mp4

import (
   "bytes"
   "encoding/hex"
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

// TestDecryption is a table-driven test that decrypts all provided samples.
func TestDecryption(t *testing.T) {
   const testDataPrefix = "../../testdata/"
   const outputDir = "test_output"

   if err := os.MkdirAll(outputDir, 0755); err != nil {
      t.Fatalf("Could not create output directory: %v", err)
   }

   for _, test := range senc_tests {
      t.Run(test.out, func(t *testing.T) {
         initFilePath := filepath.Join(testDataPrefix, test.initial)
         segmentFilePath := filepath.Join(testDataPrefix, test.segment)

         initData, err := os.ReadFile(initFilePath)
         if err != nil {
            t.Skipf("Skipping: could not read init file: %s", initFilePath)
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

         segmentData, err := os.ReadFile(segmentFilePath)
         if err != nil {
            t.Skipf("Skipping: could not read segment file: %s", segmentFilePath)
         }
         parsedSegment, err := ParseFile(segmentData)
         if err != nil {
            t.Fatalf("Failed to parse segment file: %v", err)
         }
         var moof *MoofBox
         var mdat *MdatBox
         for i := range parsedSegment {
            if parsedSegment[i].Moof != nil {
               moof = parsedSegment[i].Moof
            }
            if parsedSegment[i].Mdat != nil {
               mdat = parsedSegment[i].Mdat
            }
         }
         if moof == nil || mdat == nil {
            t.Fatal("Could not find 'moof' and/or 'mdat' box in segment.")
         }

         trak := moov.GetTrakByTrackID(1)
         if trak == nil {
            t.Fatal("Could not find track 1 in moov box.")
         }

         mdhd := trak.GetMdhd()
         if mdhd == nil {
            t.Fatal("Could not find mdhd box to calculate bandwidth.")
         }
         for _, moofChild := range moof.Children {
            if traf := moofChild.Traf; traf != nil {
               bandwidth, err := traf.GetBandwidth(mdhd.Timescale)
               if err != nil {
                  t.Errorf("Failed to calculate bandwidth: %v", err)
               } else {
                  t.Logf("Calculated Bandwidth: %d bps (%.2f kbps)", bandwidth, float64(bandwidth)/1000.0)
               }
            }
         }

         var decryptedPayload []byte
         tenc := trak.GetTenc()
         if tenc != nil {
            kidBytes := tenc.DefaultKID[:]
            keyBytes, err := hex.DecodeString(test.key)
            if err != nil {
               t.Fatalf("Failed to decode test key from hex: %v", err)
            }
            keys := make(KeyMap)
            if err := keys.AddKey(kidBytes, keyBytes); err != nil {
               t.Fatalf("Failed to add key to KeyMap: %v", err)
            }
            payload, err := keys.Decrypt(moof, mdat.Payload, moov)
            if err != nil {
               t.Fatalf("Decryption failed: %v", err)
            }
            decryptedPayload = payload
         } else {
            decryptedPayload = mdat.Payload
         }

         if err := moov.RemoveEncryption(); err != nil {
            t.Logf("Note: removeEncryption returned an error (likely expected for clear content): %v", err)
         }
         moov.RemoveDRM()
         moof.RemoveDRM()
         trak.RemoveEdts()

         var finalMP4Data bytes.Buffer
         for _, box := range parsedInit {
            if box.Moov != nil {
               finalMP4Data.Write(moov.Encode())
            } else {
               finalMP4Data.Write(box.Encode())
            }
         }
         finalMP4Data.Write(moof.Encode())

         newMdat := MdatBox{
            Header:  BoxHeader{Type: [4]byte{'m', 'd', 'a', 't'}},
            Payload: decryptedPayload,
         }
         finalMP4Data.Write(newMdat.Encode())

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

// TestPsshParsing verifies that the pssh box is parsed correctly.
func TestPsshParsing(t *testing.T) {
   psshFilePath := "../../testdata/roku-avc1/index_video_8_0_init.mp4"
   psshData, err := os.ReadFile(psshFilePath)
   if err != nil {
      t.Skipf("Skipping pssh test: could not read file: %s", psshFilePath)
   }

   parsed, err := ParseFile(psshData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   var psshBoxes []*PsshBox
   for i := range parsed {
      if parsed[i].Moov != nil {
         for j := range parsed[i].Moov.Children {
            if parsed[i].Moov.Children[j].Pssh != nil {
               psshBoxes = append(psshBoxes, parsed[i].Moov.Children[j].Pssh)
            }
         }
      }
   }

   if len(psshBoxes) != 2 {
      t.Fatalf("Expected to find 2 pssh boxes, but found %d", len(psshBoxes))
   }

   // Known SystemIDs
   widevineID, _ := hex.DecodeString("edef8ba979d64acea3c827dcd51d21ed")
   playreadyID, _ := hex.DecodeString("9a04f07998404286ab92e65be0885f95")

   foundWidevine := false
   foundPlayready := false

   for _, pssh := range psshBoxes {
      if bytes.Equal(pssh.SystemID[:], widevineID) {
         foundWidevine = true
         if len(pssh.Data) == 0 {
            t.Error("Widevine pssh box has empty Data field")
         }
         t.Logf("Found Widevine pssh box with Data length: %d", len(pssh.Data))
      }
      if bytes.Equal(pssh.SystemID[:], playreadyID) {
         foundPlayready = true
         if len(pssh.Data) == 0 {
            t.Error("PlayReady pssh box has empty Data field")
         }
         t.Logf("Found PlayReady pssh box with Data length: %d", len(pssh.Data))
      }
   }

   if !foundWidevine {
      t.Error("Did not find Widevine pssh box")
   }
   if !foundPlayready {
      t.Error("Did not find PlayReady pssh box")
   }
}
