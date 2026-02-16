package sofia

import (
   "bytes"
   "encoding/hex"
   "os"
   "path/filepath"
   "testing"
)

// TestTencBox_Parsing specifically validates the 'tenc' box parsing logic.
// It reads the init segment, navigates the box structure to the 'tenc' box,
// and asserts that the DefaultKID is parsed correctly.
func TestTencBox_Parsing(t *testing.T) {
   // 1. Define the correct, expected DefaultKID.
   expectedKIDString := "3c1863995f93b82bce88bace3a1aa67a"
   expectedKID, err := hex.DecodeString(expectedKIDString)
   if err != nil {
      t.Fatalf("Internal test error: failed to decode expected KID hex string: %v", err)
   }

   // 2. Read the test MP4 file.
   initSegment, err := os.ReadFile(filepath.Join("testdata", "init.mp4"))
   if err != nil {
      t.Fatalf("Failed to read test file 'testdata/init.mp4': %v", err)
   }

   // 3. Parse the file to get the box structure.
   boxes, err := Parse(initSegment)
   if err != nil {
      t.Fatalf("Failed to parse init segment: %v", err)
   }

   // 4. Navigate through the hierarchy to find the 'tenc' box.
   // moov -> trak -> mdia -> minf -> stbl -> stsd -> (enc*) -> sinf -> schi -> tenc
   moov, ok := FindMoov(boxes)
   if !ok {
      t.Fatal("'moov' box not found")
   }
   if len(moov.Trak) == 0 {
      t.Fatal("'trak' box not found in 'moov'")
   }
   trak := moov.Trak[0]
   if trak.Mdia == nil || trak.Mdia.Minf == nil || trak.Mdia.Minf.Stbl == nil || trak.Mdia.Minf.Stbl.Stsd == nil {
      t.Fatal("Incomplete track structure, could not find 'stsd' box")
   }
   sinf, _, ok := trak.Mdia.Minf.Stbl.Stsd.Sinf()
   if !ok {
      t.Fatal("'sinf' box not found")
   }
   if sinf.Schi == nil {
      t.Fatal("'schi' box not found in 'sinf'")
   }
   if sinf.Schi.Tenc == nil {
      t.Fatal("'tenc' box not found in 'schi'")
   }
   tencBox := sinf.Schi.Tenc

   // 5. ASSERTION: This is the actual test. Compare the parsed KID with the expected one.
   parsedKID := tencBox.DefaultKID[:]
   if !bytes.Equal(parsedKID, expectedKID) {
      t.Errorf("DefaultKID was parsed incorrectly!\n  Expected: %s\n  Got:      %s",
         hex.EncodeToString(expectedKID), hex.EncodeToString(parsedKID))
   } else {
      t.Logf("OK: DefaultKID parsed correctly as %s", hex.EncodeToString(parsedKID))
   }
}
