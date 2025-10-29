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

// TestGetBandwidth demonstrates how to parse the necessary boxes to use the GetBandwidth method.
func TestGetBandwidth(t *testing.T) {
   traf, timescale := setupTrafTest(t)

   // Call GetBandwidth with the timescale.
   bandwidth, err := traf.GetBandwidth(timescale)
   if err != nil {
      t.Fatalf("GetBandwidth failed: %v", err)
   }

   // Verify the result is plausible.
   if bandwidth == 0 {
      t.Error("Expected a non-zero bandwidth, but got 0.")
   }

   t.Logf("Successfully calculated bandwidth: %d bps (%.2f kbps)", bandwidth, float64(bandwidth)/1000.0)
}

// TestGetTotalDuration verifies the calculation of a traf's total duration.
func TestGetTotalDuration(t *testing.T) {
   traf, timescale := setupTrafTest(t)

   duration, err := traf.GetTotalDuration()
   if err != nil {
      t.Fatalf("GetTotalDuration failed: %v", err)
   }

   // Verify the result is plausible.
   if duration == 0 {
      t.Error("Expected a non-zero duration, but got 0.")
   }

   // Optional: Calculate and log the duration in seconds for readability.
   durationInSeconds := float64(duration) / float64(timescale)
   t.Logf("Successfully calculated total duration: %d (in timescale units)", duration)
   t.Logf("Timescale: %d units per second", timescale)
   t.Logf("Calculated duration: %.2f seconds", durationInSeconds)
}
