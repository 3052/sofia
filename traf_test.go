package sofia

import (
   "os"
   "path/filepath"
   "testing"
)

// TestGetBandwidth demonstrates how to parse the necessary boxes to use the GetBandwidth method.
func TestGetBandwidth(t *testing.T) {
   const testDataPrefix = "testdata/"
   initFilePath := filepath.Join(testDataPrefix, "roku-avc1/index_video_8_0_init.mp4")
   segmentFilePath := filepath.Join(testDataPrefix, "roku-avc1/index_video_8_0_1.mp4")

   // 1. First, we need the 'timescale' from the initialization segment's 'mdhd' box.
   initData, err := os.ReadFile(initFilePath)
   if err != nil {
      t.Fatalf("Could not read init file for bandwidth test: %v", err)
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

   trak := moov.GetTrak()
   if trak == nil {
      t.Fatal("Could not find 'trak' in moov.")
   }
   mdhd := trak.GetMdhd()
   if mdhd == nil {
      t.Fatal("Could not find 'mdhd' in trak to get timescale.")
   }
   timescale := mdhd.Timescale
   if timescale == 0 {
      t.Fatal("Parsed timescale is zero.")
   }

   t.Logf("Found timescale: %d", timescale)

   // 2. Now, get the 'traf' box from the media segment.
   segmentData, err := os.ReadFile(segmentFilePath)
   if err != nil {
      t.Fatalf("Could not read segment file for bandwidth test: %v", err)
   }
   parsedSegment, err := ParseFile(segmentData)
   if err != nil {
      t.Fatalf("Failed to parse segment file: %v", err)
   }

   var traf *TrafBox
   for _, box := range parsedSegment {
      if box.Moof != nil {
         // A moof contains one or more traf boxes. For this test, we'll just grab the first one.
         for _, child := range box.Moof.Children {
            if child.Traf != nil {
               traf = child.Traf
               break
            }
         }
      }
      if traf != nil {
         break
      }
   }

   if traf == nil {
      t.Fatal("Could not find 'traf' box in segment file.")
   }

   // 3. Call GetBandwidth with the timescale.
   bandwidth, err := traf.GetBandwidth(timescale)
   if err != nil {
      t.Fatalf("GetBandwidth failed: %v", err)
   }

   // 4. Verify the result is plausible.
   if bandwidth == 0 {
      t.Error("Expected a non-zero bandwidth, but got 0.")
   }

   t.Logf("Successfully calculated bandwidth: %d bps (%.2f kbps)", bandwidth, float64(bandwidth)/1000.0)
}
