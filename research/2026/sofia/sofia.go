package main

import (
   "41.neocities.org/sofia" // Using the import path as instructed.
   "crypto/aes"
   "encoding/hex"
   "fmt"
   "log"
   "os"
   "path/filepath"
   "sort"
)

const (
   keyHex        = "f67eefb3597372a25974b59805545e0f"
   testdataDir   = "testdata"
   initSegment   = "_init.mp4"
   mediaSegments = "*.m4s"
)

func main() {
   log.SetFlags(0)

   key, err := hex.DecodeString(keyHex)
   if err != nil {
      log.Fatalf("FATAL: Invalid key: %v", err)
   }
   block, err := aes.NewCipher(key)
   if err != nil {
      log.Fatalf("FATAL: Could not create AES cipher: %v", err)
   }

   initPath := filepath.Join(testdataDir, initSegment)
   initData, err := os.ReadFile(initPath)
   if err != nil {
      log.Fatalf("FATAL: Could not read init file '%s': %v", initPath, err)
   }

   initBoxes, _ := sofia.Parse(initData)
   moov, _ := sofia.FindMoov(initBoxes)
   outputFilename := "video.mp4"
   if moov != nil && moov.IsAudio() {
      outputFilename = "audio.mp4"
   }

   outFile, err := os.Create(outputFilename)
   if err != nil {
      log.Fatalf("FATAL: Could not create output file '%s': %v", outputFilename, err)
   }
   defer outFile.Close()

   remuxer := sofia.Remuxer{Writer: outFile}
   remuxer.OnSample = func(sample []byte, encInfo *sofia.SampleEncryptionInfo) {
      sofia.DecryptSample(sample, encInfo, block)
   }

   if err := remuxer.Initialize(initData); err != nil {
      log.Fatalf("FATAL: Failed to initialize remuxer: %v", err)
   }
   log.Printf("Writing to %s...", outputFilename)

   segmentFiles, _ := filepath.Glob(filepath.Join(testdataDir, mediaSegments))
   sort.Strings(segmentFiles)

   for _, segFile := range segmentFiles {
      log.Printf("  Processing %s...", filepath.Base(segFile))
      segmentData, _ := os.ReadFile(segFile)
      if err := remuxer.AddSegment(segmentData); err != nil {
         log.Printf("  WARN: Could not process segment '%s', skipping: %v", segFile, err)
      }
   }

   if err := remuxer.Finish(); err != nil {
      log.Fatalf("FATAL: Failed to finalize remuxing: %v", err)
   }

   fmt.Printf("\n✅ Success. Output: %s\n", outputFilename)
}
