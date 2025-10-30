package sofia

import (
   "os"
   "testing"
)

// TestSidxParsing verifies that the sidx box is parsed correctly.
func TestSidxParsing(t *testing.T) {
   sidxFilePath := "testdata/roku-avc1/index_video_8_0_1.mp4"
   sidxData, err := os.ReadFile(sidxFilePath)
   if err != nil {
      t.Fatalf("Could not read file: %s, error: %v", sidxFilePath, err)
   }
   parsed, err := Parse(sidxData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   sidx, ok := FindSidx(parsed)
   if !ok {
      t.Fatal("sidx box not found in file")
   }

   // Based on the mp4dump for this file:
   // - sidx version=1
   // - reference[1]: type=0 size=11433
   expectedVersion := byte(1)
   if sidx.Version != expectedVersion {
      t.Errorf("incorrect sidx version: got %d, want %d", sidx.Version, expectedVersion)
   }
   expectedRefCount := 1
   if len(sidx.References) != expectedRefCount {
      t.Fatalf("incorrect reference count: got %d, want %d", len(sidx.References), expectedRefCount)
   }
   expectedSize := uint32(11433)
   if sidx.References[0].ReferencedSize != expectedSize {
      t.Errorf("incorrect referenced_size: got %d, want %d", sidx.References[0].ReferencedSize, expectedSize)
   }
   t.Logf("Successfully parsed sidx box with referenced_size: %d", sidx.References[0].ReferencedSize)
}
