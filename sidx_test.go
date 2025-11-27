package sofia

import (
   "fmt"
   "io"
   "os"
   "path/filepath"
   "testing"
   "time"
)

const sidx_folder = "ignore"

func TestDecryptedConcatStrictSidx(t *testing.T) {
   // 1. Setup Input Files
   files, err := filepath.Glob(filepath.Join(sidx_folder, "*.m4s"))
   if err != nil {
      t.Fatalf("Failed to glob m4s files: %v", err)
   }
   outputName := fmt.Sprint(time.Now().Unix()) + ".mp4"
   if len(files) < 2 {
      t.Skip("Need at least 2 segments (init + media) to run test")
   }
   initSeg := files[0]
   mediaSegs := files[1:]
   // 2. Create Output File
   outPath := filepath.Join(sidx_folder, outputName)
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
   _, err = outFile.Write(initData)
   if err != nil {
      t.Fatal(err)
   }
   initBoxes, err := Parse(initData)
   if err != nil {
      t.Fatalf("Failed to parse init segment: %v", err)
   }
   var timescale uint32
   for _, b := range initBoxes {
      if b.Moov != nil {
         if trak, ok := b.Moov.Trak(); ok {
            if mdia, ok := trak.Mdia(); ok {
               if mdhd, ok := mdia.Mdhd(); ok {
                  timescale = mdhd.Timescale
               }
            }
         }
      }
   }
   if timescale == 0 {
      t.Fatal("timescale not found in init segment")
   }
   initEndOffset, _ := outFile.Seek(0, io.SeekCurrent)
   // 4. Write Placeholder Sidx (Strict Size Calculation)
   var dummySidx SidxBox
   // Add one dummy reference per media segment
   for range mediaSegs {
      dummySidx.AddReference(0, 0)
   }
   placeholderBytes := dummySidx.Encode()
   placeholderSize := len(placeholderBytes)
   if _, err := outFile.Write(placeholderBytes); err != nil {
      t.Fatalf("Failed to write sidx placeholder: %v", err)
   }
   // 5. Process Media Segments
   realSidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      ReferenceID: 1,
      Timescale:   timescale,
   }
   for _, segPath := range mediaSegs {
      t.Log(segPath)
      segData, err := os.ReadFile(segPath)
      if err != nil {
         t.Fatalf("Failed to read segment %s: %v", segPath, err)
      }
      boxes, err := Parse(segData)
      if err != nil {
         t.Fatalf("Failed to parse segment %s: %v", segPath, err)
      }
      // Calculate Duration & Sanitize
      var segDuration uint64
      for _, b := range boxes {
         if b.Moof != nil {
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
      _, err = outFile.Write(segData)
      if err != nil {
         t.Fatal(err)
      }
      // Add Real Reference (one per segment)
      err = realSidx.AddReference(uint32(len(segData)), uint32(segDuration))
      if err != nil {
         t.Fatalf("Failed to add reference for segment %s: %v", segPath, err)
      }
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
