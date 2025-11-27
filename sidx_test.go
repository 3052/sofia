package sofia

import (
   "encoding/binary"
   "io"
   "os"
   "path/filepath"
   "sort"
   "testing"
)

// TestSinglePassConcat performs a one-pass read/write concatenation.
// 1. Writes Init segment.
// 2. Writes a placeholder Sidx (allocated for max size).
// 3. Streams Media segments to output while extracting duration (no double read).
// 4. Seeks back and updates Sidx.
func TestSinglePassConcat(t *testing.T) {
   // 1. Setup Input Files
   files, err := filepath.Glob(filepath.Join("testdata", "*.mp4"))
   if err != nil {
      t.Fatalf("Failed to glob mp4 files: %v", err)
   }
   outputName := "output_concatenated.mp4"
   var inputs []string
   for _, f := range files {
      if filepath.Base(f) != outputName {
         inputs = append(inputs, f)
      }
   }
   sort.Strings(inputs)

   if len(inputs) < 2 {
      t.Skip("Need at least 2 segments (init + media) to run test")
   }

   initSeg := inputs[0]
   mediaSegs := inputs[1:]

   // 2. Open Output File
   outPath := filepath.Join("testdata", outputName)
   outFile, err := os.Create(outPath)
   if err != nil {
      t.Fatalf("Failed to create output file: %v", err)
   }
   defer outFile.Close()

   // 3. Write Init Segment
   initData, err := os.ReadFile(initSeg)
   if err != nil {
      t.Fatalf("Failed to read init segment: %v", err)
   }
   if _, err := outFile.Write(initData); err != nil {
      t.Fatalf("Failed to write init segment: %v", err)
   }

   // Extract timescale from Init segment for the Sidx
   parsedInit, _ := Parse(initData)
   var timescale uint32 = 90000
   if moov, ok := FindMoov(parsedInit); ok {
      if trak, ok := moov.Trak(); ok {
         if mdia, ok := trak.Mdia(); ok {
            if mdhd, ok := mdia.Mdhd(); ok {
               timescale = mdhd.Timescale
            }
         }
      }
   }

   // 4. Write Placeholder Sidx
   // We construct a Sidx with dummy references to calculate the size.
   dummySidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     0,
      ReferenceID: 1,
      Timescale:   timescale,
   }
   // Add dummy references equal to the number of segments
   for range mediaSegs {
      // Use max values to ensure size reservation is sufficient (though MP4 integers are fixed width)
      dummySidx.AddReference(0xFFFFFFFF, 0xFFFFFFFF, true, 0, 0)
   }

   placeholderBytes := dummySidx.Encode()
   placeholderSize := len(placeholderBytes)
   sidxStartOffset, _ := outFile.Seek(0, io.SeekCurrent)

   // Write the placeholder to disk
   if _, err := outFile.Write(placeholderBytes); err != nil {
      t.Fatalf("Failed to write sidx placeholder: %v", err)
   }

   // 5. Process Segments (Stream & Parse)
   // We will populate the real Sidx struct as we go.
   realSidx := &SidxBox{
      Header:      BoxHeader{Type: [4]byte{'s', 'i', 'd', 'x'}},
      Version:     0,
      ReferenceID: 1,
      Timescale:   timescale,
      // FirstOffset must be the size of the Sidx box itself, as it points to the byte
      // immediately following the Sidx box.
      FirstOffset: uint64(placeholderSize),
   }

   for _, segPath := range mediaSegs {
      size, duration, err := copySegmentAndExtractDuration(segPath, outFile)
      if err != nil {
         t.Fatalf("Failed to process segment %s: %v", segPath, err)
      }
      // Add real reference
      realSidx.AddReference(size, duration, true, 1, 0)
   }

   // 6. Overwrite Sidx
   realBytes := realSidx.Encode()

   // Safety check: Calculate padding if the real sidx is smaller than the placeholder
   // (This shouldn't happen with fixed-width fields, but good for robustness)
   if len(realBytes) > placeholderSize {
      t.Fatalf("Real Sidx size (%d) exceeds placeholder size (%d)", len(realBytes), placeholderSize)
   }
   padding := placeholderSize - len(realBytes)

   // Seek back to Sidx start
   if _, err := outFile.Seek(sidxStartOffset, io.SeekStart); err != nil {
      t.Fatalf("Failed to seek to sidx offset: %v", err)
   }

   // Write real Sidx
   if _, err := outFile.Write(realBytes); err != nil {
      t.Fatalf("Failed to overwrite sidx: %v", err)
   }

   // Fill gap with 'free' box if needed
   if padding > 0 {
      if padding < 8 {
         // MP4 boxes need at least 8 bytes header.
         // In practice, sidx size is deterministic so this branch is unlikely.
         // If it happens, we'd need to pad inside the Sidx reserved fields if possible,
         // or just accept a small garbage gap (which is technically invalid ISO BMFF but often ignored).
         // For this test, we assume standard behavior.
         t.Logf("Warning: Padding %d is less than 8 bytes, cannot write standard free box.", padding)
         // Fill with zeros (skip)
         zeros := make([]byte, padding)
         outFile.Write(zeros)
      } else {
         freeHeader := make([]byte, 8)
         binary.BigEndian.PutUint32(freeHeader[0:4], uint32(padding))
         copy(freeHeader[4:8], []byte("free"))

         // Write header
         outFile.Write(freeHeader)

         // Write payload (padding - 8)
         if padding > 8 {
            filler := make([]byte, padding-8)
            outFile.Write(filler)
         }
      }
   }

   t.Logf("Single-pass concatenation complete: %s", outPath)
}

// copySegmentAndExtractDuration copies data from the input file to the writer
// while parsing 'moof' atoms to calculate the segment duration.
// It returns total bytes copied and the duration.
func copySegmentAndExtractDuration(path string, w io.Writer) (uint32, uint32, error) {
   f, err := os.Open(path)
   if err != nil {
      return 0, 0, err
   }
   defer f.Close()

   var totalDuration uint64
   var totalBytes int64
   header := make([]byte, 8)

   for {
      // Read Atom Header
      n, err := io.ReadFull(f, header)
      if err == io.EOF {
         break
      }
      if err != nil {
         return 0, 0, err
      }

      // Write Header to Output
      if _, err := w.Write(header); err != nil {
         return 0, 0, err
      }
      totalBytes += int64(n)

      size := binary.BigEndian.Uint32(header[0:4])
      typ := string(header[4:8])

      // Basic handling for box size
      if size < 8 {
         // size 0 means "rest of file"
         if size == 0 {
            copied, err := io.Copy(w, f)
            if err != nil {
               return 0, 0, err
            }
            totalBytes += copied
            break
         }
         // size 1 (Extended) not implemented for this snippet
         return 0, 0, io.ErrUnexpectedEOF
      }

      payloadSize := int64(size) - 8

      if typ == "moof" {
         // We need to parse 'moof' to get duration.
         // Read the payload into memory.
         payload := make([]byte, payloadSize)
         if _, err := io.ReadFull(f, payload); err != nil {
            return 0, 0, err
         }

         // Write payload to Output
         if _, err := w.Write(payload); err != nil {
            return 0, 0, err
         }
         totalBytes += payloadSize

         // Reconstruct full box data for the sofia parser
         fullBox := make([]byte, size)
         copy(fullBox[:8], header)
         copy(fullBox[8:], payload)

         var moof MoofBox
         if err := moof.Parse(fullBox); err != nil {
            return 0, 0, err
         }

         if traf, ok := moof.Traf(); ok {
            _, d, err := traf.Totals()
            if err != nil {
               return 0, 0, err
            }
            totalDuration += d
         }
      } else {
         // For 'mdat' and others, stream directly without buffering everything
         copied, err := io.CopyN(w, f, payloadSize)
         if err != nil {
            return 0, 0, err
         }
         totalBytes += copied
      }
   }

   return uint32(totalBytes), uint32(totalDuration), nil
}
