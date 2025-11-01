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
   testFilePath := filepath.Join("testdata", "hulu-avc1/map.mp4")
   expectedKidHex := "077ad79156b4442a9b44e17692764a4a"
   expectedKid, err := hex.DecodeString(expectedKidHex)
   if err != nil {
      t.Fatalf("Internal test error: failed to decode expected KID hex: %v", err)
   }

   initData, err := os.ReadFile(testFilePath)
   if err != nil {
      t.Fatalf("Could not read test file %s: %v", testFilePath, err)
   }
   parsedInit, err := Parse(initData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   // Navigate to the 'tenc' box, checking for existence at each step.
   moov, ok := FindMoov(parsedInit)
   if !ok {
      t.Fatal(&Missing{Child: "moov"})
   }
   trak, ok := moov.Trak()
   if !ok {
      t.Fatal(&Missing{Child: "trak"})
   }
   mdia, ok := trak.Mdia()
   if !ok {
      t.Fatal(&Missing{Child: "mdia"})
   }
   minf, ok := mdia.Minf()
   if !ok {
      t.Fatal(&Missing{Child: "minf"})
   }
   stbl, ok := minf.Stbl()
   if !ok {
      t.Fatal(&Missing{Child: "stbl"})
   }
   stsd, ok := stbl.Stsd()
   if !ok {
      t.Fatal(&Missing{Child: "stsd"})
   }
   sinf, _, ok := stsd.Sinf()
   if !ok {
      t.Fatal(&Missing{Child: "sinf"})
   }
   schi, ok := sinf.Schi()
   if !ok {
      t.Fatal(&Missing{Child: "schi"})
   }
   tenc, ok := schi.Tenc()
   if !ok {
      t.Fatal(&Missing{Child: "tenc"})
   }

   // The actual test: Compare the parsed KID with the expected KID.
   parsedKid := tenc.DefaultKID[:] // Convert array to slice for comparison.
   if !bytes.Equal(parsedKid, expectedKid) {
      t.Errorf("DefaultKID mismatch:\n got: %x\nwant: %x", parsedKid, expectedKid)
   } else {
      t.Logf("Successfully verified correct KID parsing: %x", parsedKid)
   }
}
