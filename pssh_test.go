package sofia

import (
   "encoding/hex"
   "os"
   "testing"
)

func TestPsshParsing(t *testing.T) {
   psshFilePath := "testdata/roku-avc1/index_video_8_0_init.mp4"
   psshData, err := os.ReadFile(psshFilePath)
   if err != nil {
      t.Fatalf("Could not read file: %s, error: %v", psshFilePath, err)
   }
   parsed, err := ParseFile(psshData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   moov, ok := FindMoov(parsed)
   if !ok {
      t.Fatal("Could not find 'moov' box in parsed file.")
   }

   psshBoxes := moov.AllPssh()
   if len(psshBoxes) < 2 {
      t.Fatalf("Expected to find at least 2 pssh boxes, but found %d", len(psshBoxes))
   }

   widevineID, _ := hex.DecodeString("edef8ba979d64acea3c827dcd51d21ed")
   playreadyID, _ := hex.DecodeString("9a04f07998404286ab92e65be0885f95")

   // Test for Widevine pssh box.
   widevineBox, ok := FindPssh(psshBoxes, widevineID)
   if !ok {
      t.Error("Did not find Widevine pssh box")
   } else {
      if len(widevineBox.Data) == 0 {
         t.Error("Widevine pssh box has an empty Data field")
      }
      t.Logf("Found Widevine pssh box with Data length: %d", len(widevineBox.Data))
   }

   // Test for PlayReady pssh box.
   playreadyBox, ok := FindPssh(psshBoxes, playreadyID)
   if !ok {
      t.Error("Did not find PlayReady pssh box")
   } else {
      if len(playreadyBox.Data) == 0 {
         t.Error("PlayReady pssh box has an empty Data field")
      }
      t.Logf("Found PlayReady pssh box with Data length: %d", len(playreadyBox.Data))
   }
}
