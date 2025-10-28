package sofia

import (
   "bytes"
   "encoding/hex"
   "os"
   "testing"
)

func TestPsshParsing(t *testing.T) {
   psshFilePath := "../../testdata/roku-avc1/index_video_8_0_init.mp4"
   psshData, err := os.ReadFile(psshFilePath)
   if err != nil {
      t.Skipf("Skipping pssh test: could not read file: %s", psshFilePath)
   }

   parsed, err := ParseFile(psshData)
   if err != nil {
      t.Fatalf("Failed to parse file: %v", err)
   }

   var psshBoxes []*PsshBox
   for i := range parsed {
      if parsed[i].Moov != nil {
         for j := range parsed[i].Moov.Children {
            if parsed[i].Moov.Children[j].Pssh != nil {
               psshBoxes = append(psshBoxes, parsed[i].Moov.Children[j].Pssh)
            }
         }
      }
   }

   if len(psshBoxes) != 2 {
      t.Fatalf("Expected to find 2 pssh boxes, but found %d", len(psshBoxes))
   }

   // Known SystemIDs
   widevineID, _ := hex.DecodeString("edef8ba979d64acea3c827dcd51d21ed")
   playreadyID, _ := hex.DecodeString("9a04f07998404286ab92e65be0885f95")

   foundWidevine := false
   foundPlayready := false

   for _, pssh := range psshBoxes {
      if bytes.Equal(pssh.SystemID[:], widevineID) {
         foundWidevine = true
         if len(pssh.Data) == 0 {
            t.Error("Widevine pssh box has empty Data field")
         }
         t.Logf("Found Widevine pssh box with Data length: %d", len(pssh.Data))
      }
      if bytes.Equal(pssh.SystemID[:], playreadyID) {
         foundPlayready = true
         if len(pssh.Data) == 0 {
            t.Error("PlayReady pssh box has empty Data field")
         }
         t.Logf("Found PlayReady pssh box with Data length: %d", len(pssh.Data))
      }
   }

   if !foundWidevine {
      t.Error("Did not find Widevine pssh box")
   }
   if !foundPlayready {
      t.Error("Did not find PlayReady pssh box")
   }
}
