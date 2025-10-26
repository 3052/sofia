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
            return // Skip empty files
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

func TestDecryption(t *testing.T) {
   const placeholderKey = "1ba08384626f9523e37b9db17f44da2b"
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
   for i := range parsedInit {
      if parsedInit[i].Moov != nil {
         moov = parsedInit[i].Moov
      }
   }
   if moov == nil {
      t.Fatal("Could not find 'moov' box in init file.")
   }

   trak := moov.GetTrakByTrackID(1)
   if trak == nil {
      t.Fatal("Could not find video track in moov box.")
   }
   tenc := trak.GetTenc()
   if tenc == nil {
      t.Fatal("Could not find 'tenc' box. Is the content actually encrypted?")
   }
   kidFromFile := hex.EncodeToString(tenc.DefaultKID[:])

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
   for i := range parsedSegment {
      if parsedSegment[i].Moof != nil {
         moof = parsedSegment[i].Moof
      }
      if parsedSegment[i].Mdat != nil {
         mdat = parsedSegment[i].Mdat
      }
   }
   if moof == nil || mdat == nil {
      t.Fatal("Could not find 'moof' and/or 'mdat' box.")
   }

   decrypter := NewDecrypter()
   if err := decrypter.AddKey(kidFromFile, placeholderKey); err != nil {
      t.Fatalf("Failed to add key: %v", err)
   }

   decryptedPayload, err := decrypter.Decrypt(moof, mdat.Data[8:], moov)
   if err != nil {
      t.Fatalf("Decryption failed: %v", err)
   }

   if err := removeEncryption(moov); err != nil {
      t.Fatalf("Failed to replace 'encv' with 'avc1': %v", err)
   }

   removeDRM(moov, moof)

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

   outputFilePath := "roku-avc1.mp4"
   if err := os.WriteFile(outputFilePath, finalMP4Data.Bytes(), 0644); err != nil {
      t.Fatalf("Failed to write final MP4 file: %v", err)
   }

   if bytes.Contains(finalMP4Data.Bytes(), []byte("pssh")) {
      t.Error("Final MP4 contains 'pssh' box; removal failed.")
   }
   if bytes.Contains(finalMP4Data.Bytes(), []byte("sinf")) {
      t.Error("Final MP4 contains 'sinf' box; removal failed.")
   }
}
