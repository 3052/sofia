package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "path/filepath"
   "sort"
   "testing"
)

func TestDecryptedConcatStrictSidx(t *testing.T) {
   // 0. Prepare Key
   keyHex := "13d7c7cf295444944b627ef0ad2c1b3c"
   key, err := hex.DecodeString(keyHex)
   if err != nil {
      t.Fatalf("Invalid key hex: %v", err)
   }

   // 1. Setup Input Files
   files, err := filepath.Glob(filepath.Join("testdata", "*.mp4"))
   if err != nil {
      t.Fatalf("Failed to glob mp4 files: %v", err)
   }

   outputName := "output_decrypted_strict.mp4"
   var inputs []string
   for _, f := range files {
      base := filepath.Base(f)
      // Exclude previous output files
      if base != outputName && base != "output_concatenated.mp4" && base != "output_decrypted.mp4" {
         inputs = append(inputs, f)
      }
   }
   sort.Strings(inputs)

   if len(inputs) < 2 {
      t.Skip("Need at least 2 segments (init + media) to run test")
   }

   initSeg := inputs[0]
   mediaSegs := inputs[1:]

   // 2. Create Output File
   outPath := filepath.Join("testdata", outputName)
   outFile, err := os.Create(outPath)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer outFile.Close()

   // 3. Process Init Segment (Sanitize Only)
   initData, err := os.ReadFile(initSeg)
   if err != nil {
      t.Fatalf("Failed to read init segment: %v", err)
   }

   initBoxes, err := Parse(initData)
   if err != nil {
      t.Fatalf("Failed to parse init segment: %v", err)
   }

   var timescale uint32 = 90000

   for _, b := range initBoxes {
      if b.Moov != nil {
         if err := b.Moov.Sanitize(); err != nil {
            t.Logf("Warning sanitizing init segment: %v", err)
         }
         if trak, ok := b.Moov.Trak(); ok {
            if mdia, ok := trak.Mdia(); ok {
               if mdhd, ok := mdia.Mdhd(); ok {
                  timescale = mdhd.Timescale
               }
            }
         }
      }
      if _, err := outFile.Write(b.Encode()); err != nil {
         t.Fatalf("Failed to write init box: %v", err)
      }
   }

   initEndOffset, _ := outFile.Seek(0, io.SeekCurrent)

   // 4. Write Placeholder Sidx (Strict Size Calculation)
   // We use the exact same Version (0) and Reference Count as the final box.
   dummySidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     0,
      ReferenceID: 1,
      Timescale:   timescale,
   }

   // Add dummy references strictly equal to the number of media segments.
   for range mediaSegs {
      dummySidx.AddReference(0, 0, false, 0, 0)
   }

   placeholderBytes := dummySidx.Encode()
   placeholderSize := len(placeholderBytes)

   if _, err := outFile.Write(placeholderBytes); err != nil {
      t.Fatalf("Failed to write sidx placeholder: %v", err)
   }

   // 5. Process Media Segments
   currentFirstOffset := uint64(initEndOffset) + uint64(placeholderSize)

   realSidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     0, // Must match dummy version
      ReferenceID: 1,
      Timescale:   timescale,
      FirstOffset: currentFirstOffset,
   }

   for _, segPath := range mediaSegs {
      segData, err := os.ReadFile(segPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segPath, err)
      }

      boxes, err := Parse(segData)
      if err != nil {
         t.Fatalf("Failed to parse segment %s: %v", segPath, err)
      }

      // Decrypt
      if err := Decrypt(boxes, key); err != nil {
         t.Fatalf("Failed to decrypt segment %s: %v", segPath, err)
      }

      // Calculate Duration & Sanitize
      var segDuration uint64
      for _, b := range boxes {
         if b.Moof != nil {
            b.Moof.Sanitize()
            if traf, ok := b.Moof.Traf(); ok {
               _, d, err := traf.Totals()
               if err == nil {
                  segDuration += d
               }
            }
         }
      }

      // Write Segment
      var segOutput []byte
      for _, b := range boxes {
         segOutput = append(segOutput, b.Encode()...)
      }

      if _, err := outFile.Write(segOutput); err != nil {
         t.Fatalf("Failed to write segment %s: %v", segPath, err)
      }

      // Add Real Reference
      realSidx.AddReference(uint32(len(segOutput)), uint32(segDuration), true, 1, 0)
   }

   // 6. Overwrite Sidx (Strict Check)
   realBytes := realSidx.Encode()

   if len(realBytes) != placeholderSize {
      t.Fatalf("Sidx size mismatch! Placeholder: %d, Real: %d. Determinism failed.", placeholderSize, len(realBytes))
   }

   if _, err := outFile.Seek(initEndOffset, io.SeekStart); err != nil {
      t.Fatalf("Failed to seek to sidx position: %v", err)
   }

   if _, err := outFile.Write(realBytes); err != nil {
      t.Fatalf("Failed to overwrite sidx: %v", err)
   }

   t.Logf("Strict decrypted concatenation complete. Output: %s", outPath)
}
