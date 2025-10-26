package mp4

import (
   "bytes"
   "encoding/binary"
   "encoding/hex"
   "errors"
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

// TestRoundTrip is now a table-driven test covering all files.
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
                  return // Skip empty files
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

// TestDecryption is now a table-driven test that decrypts all provided samples.
func TestDecryption(t *testing.T) {
   const testDataPrefix = "../../testdata/"
   const outputDir = "test_output"

   if err := os.MkdirAll(outputDir, 0755); err != nil {
      t.Fatalf("Could not create output directory: %v", err)
   }

   for _, test := range senc_tests {
      t.Run(test.out, func(t *testing.T) {
         // --- Load Files ---
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

         // --- Decryption & Sanitization ---
         trak := moov.GetTrakByTrackID(1) // Assumes track ID 1
         if trak == nil {
            t.Fatal("Could not find track 1 in moov box.")
         }
         tenc := trak.GetTenc()
         if tenc == nil {
            t.Fatal("Could not find 'tenc' box. Content may not be encrypted.")
         }
         kidFromFile := hex.EncodeToString(tenc.DefaultKID[:])

         decrypter := NewDecrypter()
         if err := decrypter.AddKey(kidFromFile, test.key); err != nil {
            t.Fatalf("Failed to add key: %v", err)
         }

         decryptedPayload, err := decrypter.Decrypt(moof, mdat.Data[8:], moov)
         if err != nil {
            t.Fatalf("Decryption failed: %v", err)
         }

         if err := removeEncryption(moov); err != nil {
            t.Fatalf("Failed to replace encryption signaling: %v", err)
         }
         removeDRM(moov, moof)
         removeEdts(moov)

         // --- Assemble and Write File ---
         var finalMP4Data bytes.Buffer
         for _, box := range parsedInit {
            if box.Moov != nil {
               finalMP4Data.Write(moov.Encode())
            } else {
               finalMP4Data.Write(box.Encode())
            }
         }
         finalMP4Data.Write(moof.Encode())
         finalMP4Data.Write(createMdatBox(decryptedPayload))

         outputFilePath := filepath.Join(outputDir, test.out)
         if err := os.WriteFile(outputFilePath, finalMP4Data.Bytes(), 0644); err != nil {
            t.Fatalf("Failed to write final MP4 file: %v", err)
         }
         t.Logf("Successfully wrote decrypted file to: %s", outputFilePath)

         // --- Verification ---
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

// --- Helper Functions (unchanged) ---

func createMdatBox(payload []byte) []byte {
   size := uint32(len(payload) + 8)
   mdatBox := make([]byte, size)
   binary.BigEndian.PutUint32(mdatBox[0:4], size)
   copy(mdatBox[4:8], "mdat")
   copy(mdatBox[8:], payload)
   return mdatBox
}

func removeEncryption(moov *MoovBox) error {
   for _, trak := range moov.GetAllTraks() {
      stsd := trak.GetStsd()
      if stsd == nil {
         continue
      }
      for i, child := range stsd.Children {
         var encBoxHeader []byte
         var encChildren []interface{}
         var isVideo bool
         if child.Encv != nil {
            encBoxHeader, isVideo = child.Encv.EntryHeader, true
            for _, c := range child.Encv.Children {
               encChildren = append(encChildren, c)
            }
         } else if child.Enca != nil {
            encBoxHeader = child.Enca.EntryHeader
            for _, c := range child.Enca.Children {
               encChildren = append(encChildren, c)
            }
         } else {
            continue
         }
         var sinf *SinfBox
         if isVideo {
            for _, c := range encChildren {
               if s := c.(EncvChild).Sinf; s != nil {
                  sinf = s
                  break
               }
            }
         } else {
            for _, c := range encChildren {
               if s := c.(EncaChild).Sinf; s != nil {
                  sinf = s
                  break
               }
            }
         }
         if sinf == nil {
            return errors.New("could not find 'sinf' box")
         }
         var frma *FrmaBox
         for _, sinfChild := range sinf.Children {
            if f := sinfChild.Frma; f != nil {
               frma = f
               break
            }
         }
         if frma == nil {
            return errors.New("could not find 'frma' box")
         }
         newFormatType := frma.DataFormat
         var newContent bytes.Buffer
         newContent.Write(encBoxHeader)
         for _, c := range encChildren {
            var childSinf *SinfBox
            if isVideo {
               childSinf = c.(EncvChild).Sinf
            } else {
               childSinf = c.(EncaChild).Sinf
            }
            if childSinf == nil {
               if isVideo {
                  newContent.Write(c.(EncvChild).Raw)
               } else {
                  newContent.Write(c.(EncaChild).Raw)
               }
            }
         }
         newBoxSize := uint32(8 + newContent.Len())
         newBoxData := make([]byte, newBoxSize)
         binary.BigEndian.PutUint32(newBoxData[0:4], newBoxSize)
         copy(newBoxData[4:8], newFormatType[:])
         copy(newBoxData[8:], newContent.Bytes())
         stsd.Children[i] = StsdChild{Raw: newBoxData}
      }
   }
   return nil
}

func removeDRM(moov *MoovBox, moof *MoofBox) {
   if moov != nil {
      for i := range moov.Children {
         child := &moov.Children[i]
         if child.Pssh != nil {
            freeBoxData := make([]byte, len(child.Pssh.RawData))
            copy(freeBoxData, child.Pssh.RawData)
            copy(freeBoxData[4:8], "free")
            child.Pssh = nil
            child.Raw = freeBoxData
         }
      }
   }
   if moof != nil {
      for i := range moof.Children {
         child := &moof.Children[i]
         if child.Pssh != nil {
            freeBoxData := make([]byte, len(child.Pssh.RawData))
            copy(freeBoxData, child.Pssh.RawData)
            copy(freeBoxData[4:8], "free")
            child.Pssh = nil
            child.Raw = freeBoxData
         }
      }
   }
}

func removeEdts(moov *MoovBox) {
   if moov == nil {
      return
   }
   for _, trak := range moov.GetAllTraks() {
      for i := range trak.Children {
         child := &trak.Children[i]
         if child.Edts != nil {
            freeBoxData := make([]byte, len(child.Edts.RawData))
            copy(freeBoxData, child.Edts.RawData)
            copy(freeBoxData[4:8], "free")
            child.Edts = nil
            child.Raw = freeBoxData
         }
      }
   }
}
