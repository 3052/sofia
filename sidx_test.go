package sofia

import (
   "os"
   "path/filepath"
   "sort"
   "testing"
)

// TestBuildSidxAndConcat reads all mp4s in testdata, constructs a sidx box
// based on segments (files starting from index 1), and writes a concatenated
// stream [InitSegment][Sidx][Segment1][Segment2]... to disk.
func TestBuildSidxAndConcat(t *testing.T) {
   // 1. Read all ".mp4" files in "testdata"
   matches, err := filepath.Glob(filepath.Join("testdata", "*.mp4"))
   if err != nil {
      t.Fatalf("Failed to glob mp4 files: %v", err)
   }

   // Filter out the output file if it already exists from a previous run
   // to prevent it from being treated as an input segment.
   outputFilename := "output_concatenated.mp4"
   var inputFiles []string
   for _, m := range matches {
      if filepath.Base(m) != outputFilename {
         inputFiles = append(inputFiles, m)
      }
   }

   if len(inputFiles) == 0 {
      t.Skip("No mp4 files found in testdata to process.")
   }

   // Ensure deterministic order: Init segment first, then segments in order
   sort.Strings(inputFiles)

   // Identify Init segment and Media segments
   initPath := inputFiles[0]
   segPaths := inputFiles[1:]

   t.Logf("Init Segment: %s", initPath)
   t.Logf("Media Segments: %d files", len(segPaths))

   // Read Init Segment
   initData, err := os.ReadFile(initPath)
   if err != nil {
      t.Fatalf("Failed to read init file: %v", err)
   }

   // Parse Init Segment to extract Timescale from 'mdhd'
   initBoxes, err := Parse(initData)
   if err != nil {
      t.Fatalf("Failed to parse init segment: %v", err)
   }

   var timescale uint32
   if moov, ok := FindMoov(initBoxes); ok {
      if trak, ok := moov.Trak(); ok {
         if mdia, ok := trak.Mdia(); ok {
            if mdhd, ok := mdia.Mdhd(); ok {
               timescale = mdhd.Timescale
            }
         }
      }
   }
   // Fallback if timescale not found
   if timescale == 0 {
      t.Log("Warning: Could not find timescale in init segment, defaulting to 90000")
      timescale = 90000
   }

   // 2. Build "sidx" from segments
   // Note: FirstOffset is the distance from the anchor point (start of sidx)
   // to the first byte of the referenced data. Since we write [Sidx][Seg1]...,
   // FirstOffset will equal the size of the Sidx box itself.
   sidx := &SidxBox{
      Header:                   BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:                  0,
      ReferenceID:              1,
      Timescale:                timescale,
      EarliestPresentationTime: 0, // Assuming stream starts at 0
      FirstOffset:              0, // Will be calculated after references are added
   }

   for _, segPath := range segPaths {
      segData, err := os.ReadFile(segPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segPath, err)
      }

      // Parse segment to calculate duration
      segBoxes, err := Parse(segData)
      if err != nil {
         t.Fatalf("Failed to parse segment %s: %v", segPath, err)
      }

      var segDuration uint64
      // Aggregate duration from all 'traf' boxes in the segment
      for _, moof := range AllMoof(segBoxes) {
         if traf, ok := moof.Traf(); ok {
            _, d, err := traf.Totals()
            if err != nil {
               t.Fatalf("Failed to get totals for traf in %s: %v", segPath, err)
            }
            segDuration += d
         }
      }

      // Add reference to Sidx
      // ReferencedSize = File/Segment size in bytes
      // StartsWithSAP = true (common for DASH segments)
      // SAPType = 1, SAPDeltaTime = 0
      err = sidx.AddReference(uint32(len(segData)), uint32(segDuration), true, 1, 0)
      if err != nil {
         t.Fatalf("Failed to add reference for %s: %v", segPath, err)
      }
   }

   // Calculate and set FirstOffset.
   // 1. Encode to get the size of the sidx box.
   encodedSidx := sidx.Encode()
   sidxSize := uint64(len(encodedSidx))

   // 2. Set FirstOffset to the size of the box (pointing to byte immediately after sidx)
   sidx.FirstOffset = sidxSize

   // 3. Re-encode to bake in the correct FirstOffset
   encodedSidx = sidx.Encode()

   // Verify size consistency (size shouldn't change unless version changed, which we control)
   if uint64(len(encodedSidx)) != sidxSize {
      // In the rare case encoding changed size (e.g. variable int encoding, though MP4 uses fixed), update again.
      sidx.FirstOffset = uint64(len(encodedSidx))
      encodedSidx = sidx.Encode()
   }

   // Prepare Output File
   outPath := filepath.Join("testdata", outputFilename)
   f, err := os.Create(outPath)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer f.Close()

   // 3. Write first file (Init Segment) to disk
   _, err = f.Write(initData)
   if err != nil {
      t.Fatalf("Failed to write init data: %v", err)
   }

   // 4. Write "sidx" to disk
   _, err = f.Write(encodedSidx)
   if err != nil {
      t.Fatalf("Failed to write sidx data: %v", err)
   }

   // 5. Write remaining files (Segments) to disk
   for _, segPath := range segPaths {
      segData, err := os.ReadFile(segPath)
      if err != nil {
         t.Fatalf("Failed to read segment for writing: %v", err)
      }
      _, err = f.Write(segData)
      if err != nil {
         t.Fatalf("Failed to write segment data: %v", err)
      }
   }

   t.Logf("Created concatenated file with SIDX at: %s", outPath)
}
