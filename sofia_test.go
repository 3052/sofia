package sofia

import (
   "io"
   "math"
   "os"
   "path/filepath"
   "sort"
   "testing"
)

func TestBitrateStability(t *testing.T) {
   workDir := "ignore"
   if _, err := os.Stat(workDir); os.IsNotExist(err) {
      t.Skipf("Skipping test: directory '%s' not found", workDir)
   }

   // 1. Initialize Unfragmenter with dummy writer to parse durations
   initPath := filepath.Join(workDir, "_init.mp4")
   initData, err := os.ReadFile(initPath)
   if err != nil {
      t.Fatalf("Failed to read init segment: %v", err)
   }

   dummy := &dummyWriteSeeker{}
   u := &Unfragmenter{Writer: dummy}

   if err := u.Initialize(initData); err != nil {
      t.Fatalf("Initialize failed: %v", err)
   }

   // Get Timescale from the parsed init segment
   trak, ok := u.Moov.Trak()
   if !ok {
      t.Fatal("No trak found")
   }
   mdia, ok := trak.Mdia()
   if !ok {
      t.Fatal("No mdia found")
   }
   mdhd, ok := mdia.Mdhd()
   if !ok {
      t.Fatal("No mdhd found")
   }
   timescale := float64(mdhd.Timescale)
   if timescale == 0 {
      t.Fatal("Timescale is 0")
   }
   t.Logf("Timescale: %.0f", timescale)

   // 2. Process Segments
   globPattern := filepath.Join(workDir, "*.m4s")
   files, err := filepath.Glob(globPattern)
   if err != nil {
      t.Fatal(err)
   }
   sort.Strings(files)
   if len(files) == 0 {
      t.Skip("No segments found")
   }

   type rawStat struct {
      name string
      bits int64
      dur  float64
   }

   var rawStats []rawStat
   var totalBits int64
   var totalDuration float64

   for _, f := range files {
      data, err := os.ReadFile(f)
      if err != nil {
         t.Fatal(err)
      }

      prevSampleCount := len(u.samples)
      if err := u.AddSegment(data); err != nil {
         t.Fatalf("AddSegment failed for %s: %v", f, err)
      }

      // Calculate duration of just the added samples
      var segDurationTicks uint64
      for _, s := range u.samples[prevSampleCount:] {
         segDurationTicks += uint64(s.Duration)
      }

      segSeconds := float64(segDurationTicks) / timescale
      segBits := int64(len(data)) * 8

      totalBits += segBits
      totalDuration += segSeconds

      rawStats = append(rawStats, rawStat{
         name: filepath.Base(f),
         bits: segBits,
         dur:  segSeconds,
      })
   }

   // 3. Combined Bitrate
   combinedBitrate := 0.0
   if totalDuration > 0 {
      combinedBitrate = float64(totalBits) / totalDuration
   }
   t.Logf("Combined Bitrate: %.2f bps", combinedBitrate)

   // 4. Calculate diffs and Find Longest Run
   var currentRun []string
   var maxRun []string

   for _, r := range rawStats {
      segBitrate := 0.0
      if r.dur > 0 {
         segBitrate = float64(r.bits) / r.dur
      }

      diff := 0.0
      if combinedBitrate > 0 {
         diff = (segBitrate - combinedBitrate) / combinedBitrate
      }

      t.Logf("Segment: %s, Bitrate: %.0f, Diff: %.2f%%", r.name, segBitrate, diff*100)

      isStable := math.Abs(diff) <= 0.10

      if isStable {
         currentRun = append(currentRun, r.name)
      } else {
         if len(currentRun) > len(maxRun) {
            maxRun = make([]string, len(currentRun))
            copy(maxRun, currentRun)
         }
         currentRun = nil
      }
   }

   // Final check if run ended at the last segment
   if len(currentRun) > len(maxRun) {
      maxRun = currentRun
   }

   // 5. Print Result
   t.Logf("Longest Run (%d segments): %v", len(maxRun), maxRun)
}

// dummyWriteSeeker acts as a transparent sink to track current offset
// without writing to disk, fulfilling the io.WriteSeeker interface.
type dummyWriteSeeker struct {
   pos int64
}

func (d *dummyWriteSeeker) Write(p []byte) (int, error) {
   n := len(p)
   d.pos += int64(n)
   return n, nil
}

func (d *dummyWriteSeeker) Seek(offset int64, whence int) (int64, error) {
   switch whence {
   case io.SeekStart:
      d.pos = offset
   case io.SeekCurrent:
      d.pos += offset
   case io.SeekEnd:
      // Approximate for AddSegment usage (which usually seeks 0 from End, or just appends)
      d.pos += offset
   }
   return d.pos, nil
}
