package sofia

import (
   "crypto/aes"
   "encoding/hex"
   "fmt"
   "os"
   "testing"
   "time"
)

func TestRemuxAndDecrypt(t *testing.T) {
   // 1. Read testdata/_init.mp4 from disk
   initData, err := os.ReadFile("testdata/_init.mp4")
   if err != nil {
      t.Fatalf("Failed to read init file: %v", err)
   }

   // 2. Read testdata/0.m4s from disk
   segmentData, err := os.ReadFile("testdata/0.m4s")
   if err != nil {
      t.Fatalf("Failed to read segment file: %v", err)
   }

   // 3. Decrypt with 84b3c458e541196adcd7577b73e9f9a0
   key, err := hex.DecodeString("84b3c458e541196adcd7577b73e9f9a0")
   if err != nil {
      t.Fatalf("Failed to decode hex key: %v", err)
   }

   block, err := aes.NewCipher(key)
   if err != nil {
      t.Fatalf("Failed to create AES cipher: %v", err)
   }

   // 4. Write result to (time.Now().Unix()).mp4
   outputFileName := fmt.Sprintf("testdata/%d.mp4", time.Now().Unix())
   outFile, err := os.Create(outputFileName)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer outFile.Close()

   t.Logf("Writing decrypted output to %s", outputFileName)

   remuxer := &Remuxer{
      Writer: outFile,
      OnSample: func(sample []byte, encInfo *SampleEncryptionInfo) {
         DecryptSample(sample, encInfo, block)
      },
   }

   if err := remuxer.Initialize(initData); err != nil {
      t.Fatalf("Remuxer initialization failed: %v", err)
   }

   if err := remuxer.AddSegment(segmentData); err != nil {
      t.Fatalf("Failed to add segment: %v", err)
   }

   if err := remuxer.Finish(); err != nil {
      t.Fatalf("Remuxer finish failed: %v", err)
   }
}
