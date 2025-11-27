package sofia

import (
   "encoding/hex"
   "io"
   "os"
   "path/filepath"
   "testing"
)

var sidx = struct {
   key    string
   folder string
   write  bool
}{
   folder: "ignore",
   key:    "27736bd0d54481eab2402a879cb863c7",
   write:  true,
}

func TestDecryptedConcatStrictSidx(t *testing.T) {
   // 0. Prepare Key
   key, err := hex.DecodeString(sidx.key)
   if err != nil {
      t.Fatalf("Invalid key hex: %v", err)
   }

   // 1. Setup Input Files
   files, err := filepath.Glob(filepath.Join(sidx.folder, "*.m4s"))
   if err != nil {
      t.Fatalf("Failed to glob m4s files: %v", err)
   }

   outputName := "output_decrypted_strict.mp4"

   // Filter out potential output files
   var inputs []string
   for _, f := range files {
      if filepath.Base(f) != outputName {
         inputs = append(inputs, f)
      }
   }

   if len(inputs) < 2 {
      t.Skip("Need at least 2 segments (init + media) to run test")
   }

   initSeg := inputs[0]
   mediaSegs := inputs[1:]

   // 2. Create Output File
   outPath := filepath.Join(sidx.folder, outputName)
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

   var timescale uint32
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

   if timescale == 0 {
      t.Fatal("timescale not found in init segment")
   }

   initEndOffset, _ := outFile.Seek(0, io.SeekCurrent)

   // 4. Write Placeholder Sidx (Strict Size Calculation)
   dummySidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     1, // Version 1 for 64-bit offsets support
      ReferenceID: 1,
      Timescale:   timescale,
   }

   // Add one dummy reference per media segment
   for range mediaSegs {
      dummySidx.AddReference(0, 0, false, 0, 0)
   }

   placeholderBytes := dummySidx.Encode()
   placeholderSize := len(placeholderBytes)

   if sidx.write {
      if _, err := outFile.Write(placeholderBytes); err != nil {
         t.Fatalf("Failed to write sidx placeholder: %v", err)
      }
   }

   // 5. Process Media Segments
   realSidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     1,
      ReferenceID: 1,
      Timescale:   timescale,
      FirstOffset: 0, // Requested: 0
   }

   for _, segPath := range mediaSegs {
      // t.Log(segPath)
      segData, err := os.ReadFile(segPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segPath, err)
      }

      boxes, err := Parse(segData)
      if err != nil {
         t.Fatalf("Failed to parse segment %s: %v", segPath, err)
      }
      
      if false {
         if err := Decrypt(boxes, key); err != nil {
            t.Fatalf("Failed to decrypt segment %s: %v", segPath, err)
         }
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

      if segDuration == 0 {
         t.Fatalf("Segment %s has 0 duration", segPath)
      }

      // Write Segment
      var segOutput []byte
      for _, b := range boxes {
         segOutput = append(segOutput, b.Encode()...)
      }

      if _, err := outFile.Write(segOutput); err != nil {
         t.Fatalf("Failed to write segment %s: %v", segPath, err)
      }

      // Add Real Reference (one per segment)
      err = realSidx.AddReference(uint32(len(segOutput)), uint32(segDuration), true, 1, 0)
      if err != nil {
         t.Fatalf("Failed to add reference for segment %s: %v", segPath, err)
      }
   }

   if !sidx.write {
      return
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
