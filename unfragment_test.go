package sofia

import (
   "os"
   "path/filepath"
   "sort"
   "testing"
)

// TestUnfragmenter_RealFiles looks for real files in the "ignore" directory.
// It skips if the directory does not exist.
func TestUnfragmenter_RealFiles(t *testing.T) {
   workDir := "ignore"

   // 1. Check if "ignore" folder exists
   if _, err := os.Stat(workDir); os.IsNotExist(err) {
      t.Skipf("Skipping test: directory '%s' not found", workDir)
   }

   // 2. Open Output File
   outPath := filepath.Join(workDir, "joined_output.mp4")
   outFile, err := os.Create(outPath)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer outFile.Close()

   // 3. Initialize Unfragmenter
   unfrag := NewUnfragmenter(outFile)

   // Read init.mp4
   initPath := filepath.Join(workDir, "init.mp4")
   initData, err := os.ReadFile(initPath)
   if err != nil {
      t.Fatalf("Failed to read init segment (%s): %v", initPath, err)
   }

   if err := unfrag.Initialize(initData); err != nil {
      t.Fatalf("Unfragmenter.Initialize failed: %v", err)
   }

   // 4. Find and Sort Segments
   // Matches files like segment-1.0001.m4s, segment-1.0002.m4s
   globPattern := filepath.Join(workDir, "segment-*.m4s")
   segmentFiles, err := filepath.Glob(globPattern)
   if err != nil {
      t.Fatalf("Failed to glob segments: %v", err)
   }

   if len(segmentFiles) == 0 {
      t.Fatalf("No segment files found matching pattern: %s", globPattern)
   }

   // Sort ensures 1.0001 comes before 1.0002
   sort.Strings(segmentFiles)

   t.Logf("Found %d segments. Processing...", len(segmentFiles))

   // 5. Process Segments
   for _, segmentPath := range segmentFiles {
      // Read segment into memory
      segData, err := os.ReadFile(segmentPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segmentPath, err)
      }

      // Add to unfragmenter
      if err := unfrag.AddSegment(segData); err != nil {
         t.Fatalf("Failed to add segment %s: %v", segmentPath, err)
      }
   }

   // 6. Finish
   if err := unfrag.Finish(); err != nil {
      t.Fatalf("Unfragmenter.Finish failed: %v", err)
   }

   // 7. Report
   stat, _ := outFile.Stat()
   t.Logf("Success! Created %s (%d bytes)", outPath, stat.Size())
}
