package mp4

import (
   "bytes"
   "encoding/hex"
   "os"
   "path/filepath"
   "testing"
)

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
