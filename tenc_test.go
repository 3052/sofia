package sofia

import (
   "bytes"
   "encoding/hex"
   "os"
   "path/filepath"
   "testing"
)

// TestTencKIDParsing specifically verifies that the DefaultKID is parsed correctly from a 'tenc' box.
// This is a regression test for a bug where the parsing offset was incorrect.
func TestTencKIDParsing(t *testing.T) {
   // The file identified in the bug report.
   testFilePath := filepath.Join("testdata", "hulu-avc1/map.mp4")

   // The known correct KID for this file.
   expectedKidHex := "077ad79156b4442a9b44e17692764a4a"
   expectedKid, err := hex.DecodeString(expectedKidHex)
   if err != nil {
      t.Fatalf("Internal test error: failed to decode expected KID hex: %v", err)
   }

   // 1. Read and parse the initialization file.
   initData, err := os.ReadFile(testFilePath)
   if err != nil {
      t.Fatalf("Could not read test file %s: %v", testFilePath, err)
   }
   parsedInit, err := Parse(initData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   // 2. Navigate to the 'tenc' box.
   moov, ok := FindMoov(parsedInit)
   if !ok {
      t.Fatal("Test setup failed: could not find 'moov' box.")
   }
   trak, ok := moov.Trak()
   if !ok {
      t.Fatal("Test setup failed: could not find 'trak' box.")
   }
   tenc := trak.Tenc()
   if tenc == nil {
      t.Fatal("Test setup failed: could not find 'tenc' box.")
   }

   // 3. The actual test: Compare the parsed KID with the expected KID.
   parsedKid := tenc.DefaultKID[:] // Convert array to slice for comparison.
   if !bytes.Equal(parsedKid, expectedKid) {
      t.Errorf("DefaultKID mismatch:\n got: %x\nwant: %x", parsedKid, expectedKid)
   } else {
      t.Logf("Successfully verified correct KID parsing: %x", parsedKid)
   }
}
