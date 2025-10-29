package sofia

import (
   "os"
   "path/filepath"
   "testing"
)

// setupTrafTest is a helper to parse the necessary files for traf-related tests.
func setupTrafTest(t *testing.T) (*TrafBox, uint32) {
   t.Helper()

   const testDataPrefix = "testdata/"
   initFilePath := filepath.Join(testDataPrefix, "roku-avc1/index_video_8_0_init.mp4")
   segmentFilePath := filepath.Join(testDataPrefix, "roku-avc1/index_video_8_0_1.mp4")

   // 1. Get the 'timescale' from the initialization segment.
   initData, err := os.ReadFile(initFilePath)
   if err != nil {
      t.Fatalf("Could not read init file for test: %v", err)
   }
   parsedInit, err := ParseFile(initData)
   if err != nil {
      t.Fatalf("Failed to parse init file: %v", err)
   }
   moov := FindMoov(parsedInit)
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
   if mdhd.Timescale == 0 {
      t.Fatal("Parsed timescale is zero.")
   }

   // 2. Get the 'traf' box from the media segment.
   segmentData, err := os.ReadFile(segmentFilePath)
   if err != nil {
      t.Fatalf("Could not read segment file for test: %v", err)
   }
   parsedSegment, err := ParseFile(segmentData)
   if err != nil {
      t.Fatalf("Failed to parse segment file: %v", err)
   }
   traf := FindFirstTraf(parsedSegment)
   if traf == nil {
      t.Fatal("Could not find 'traf' box in segment file.")
   }

   return traf, mdhd.Timescale
}

// TestBandwidthCalculation demonstrates the decoupled workflow for calculating bandwidth.
func TestBandwidthCalculation(t *testing.T) {
   traf, timescale := setupTrafTest(t)
   t.Logf("Found timescale: %d", timescale)
   // 1. Get the raw totals from the TrafBox.
   totalBytes, totalDuration, err := traf.GetTotals()
   if err != nil {
      t.Fatalf("GetTotals failed: %v", err)
   }
   if totalBytes == 0 {
      t.Fatal("Expected non-zero total bytes from GetTotals.")
   }
   if totalDuration == 0 {
      t.Fatal("Expected non-zero total duration from GetTotals.")
   }
   t.Logf("Got totals: %d bytes, %d duration units", totalBytes, totalDuration)
}
