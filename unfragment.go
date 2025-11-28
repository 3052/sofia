package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
   "log"
)

// Unfragmenter handles the conversion of fragmented segments into a single file
// statefully. It keeps memory usage low by streaming payloads immediately.
type Unfragmenter struct {
   dst  io.WriteSeeker
   moov *MoovBox // The template moov from init segment

   // Track state
   samples             []sampleInfo
   chunkOffsets        []uint64 // Absolute file offsets of each segment's payload
   segmentSampleCounts []uint32 // Needed for stsc

   // Byte tracking
   mdatStartOffset int64  // Where the mdat box starts in the file
   payloadWritten  uint64 // Total bytes of media data written so far

   initialized  bool
   segmentCount int // Debugging counter
}

type sampleInfo struct {
   Size     uint32
   Duration uint32
}

// NewUnfragmenter creates a new converter writing to the provided file.
func NewUnfragmenter(dst io.WriteSeeker) *Unfragmenter {
   return &Unfragmenter{dst: dst}
}

// Initialize processes the initialization segment (init.mp4).
// It captures the moov template and writes the placeholder mdat header to output.
func (u *Unfragmenter) Initialize(initSegment []byte) error {
   if u.initialized {
      return errors.New("already initialized")
   }

   log.Println("[Unfrag] Initializing...")
   boxes, err := Parse(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init: %w", err)
   }

   // 1. Capture and validate Moov
   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found in init segment")
   }
   u.moov = moovPtr

   // Validate hierarchy
   if _, ok := u.moov.Trak(); !ok {
      return errors.New("no trak in moov")
   }

   // 2. Write mdat Header Placeholder
   u.mdatStartOffset, _ = u.dst.Seek(0, io.SeekCurrent)

   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))

   if _, err := u.dst.Write(mdatHeader); err != nil {
      return fmt.Errorf("writing mdat header: %w", err)
   }

   u.initialized = true
   log.Println("[Unfrag] Initialization complete. Ready for segments.")
   return nil
}

// AddSegment processes a single media segment (e.g., segment-1.m4s).
func (u *Unfragmenter) AddSegment(segmentData []byte) error {
   if !u.initialized {
      return errors.New("must call Initialize before AddSegment")
   }

   u.segmentCount++
   boxes, err := Parse(segmentData)
   if err != nil {
      return fmt.Errorf("parsing segment %d: %w", u.segmentCount, err)
   }

   moof := FindMoofPtr(boxes)
   mdat := FindMdatPtr(boxes)

   if moof == nil || mdat == nil {
      log.Printf("[Unfrag] Segment %d skipped: missing moof or mdat", u.segmentCount)
      return nil
   }

   // --- Step 1: Extract Metadata (in memory) ---
   traf, ok := moof.Traf()
   if !ok {
      log.Printf("[Unfrag] Segment %d skipped: valid moof but no traf", u.segmentCount)
      return nil
   }

   tfhd := traf.Tfhd()
   if tfhd == nil {
      log.Printf("[Unfrag] Segment %d skipped: missing tfhd", u.segmentCount)
      return nil
   }

   // Collect samples from ALL trun boxes in this fragment
   var newSamples []sampleInfo

   // Default values from tfhd
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize

   // Iterate over traf children to find all 'trun' boxes
   trunCount := 0
   for _, child := range traf.Children {
      if child.Trun != nil {
         trunCount++
         trun := child.Trun
         for _, s := range trun.Samples {
            si := sampleInfo{
               Duration: defDur,
               Size:     defSize,
            }
            if (trun.Flags & 0x000100) != 0 {
               si.Duration = s.Duration
            }
            if (trun.Flags & 0x000200) != 0 {
               si.Size = s.Size
            }
            newSamples = append(newSamples, si)
         }
      }
   }

   sampleCount := len(newSamples)
   if sampleCount == 0 {
      log.Printf("[Unfrag] Segment %d skipped: 0 samples found (truns found: %d)", u.segmentCount, trunCount)
      return nil
   }

   // --- Step 2: Commit to Disk/State ---

   currentPos, _ := u.dst.Seek(0, io.SeekCurrent)
   u.chunkOffsets = append(u.chunkOffsets, uint64(currentPos))

   n, err := u.dst.Write(mdat.Payload)
   if err != nil {
      return fmt.Errorf("writing payload: %w", err)
   }
   u.payloadWritten += uint64(n)

   u.samples = append(u.samples, newSamples...)
   u.segmentSampleCounts = append(u.segmentSampleCounts, uint32(sampleCount))

   log.Printf("[Unfrag] Segment %d processed: Offset=%d, Samples=%d, Bytes=%d",
      u.segmentCount, currentPos, sampleCount, n)

   return nil
}

// Finish constructs the final atoms and writes the file footer.
func (u *Unfragmenter) Finish() error {
   if !u.initialized {
      return errors.New("not initialized")
   }

   log.Println("[Unfrag] Finishing...")

   // 1. Calculate Total Duration
   var totalDuration uint64
   for _, s := range u.samples {
      totalDuration += uint64(s.Duration)
   }

   // 2. Build Sample Tables
   stts := buildStts(u.samples)
   stsz := buildStsz(u.samples)
   stsc := buildStsc(u.segmentSampleCounts)
   offsetBox := buildStco(u.chunkOffsets)

   // --- DEBUG LOGGING ---
   log.Printf("[Unfrag] Summary:")
   log.Printf("  Total Chunks (offsets): %d", len(u.chunkOffsets))
   log.Printf("  Total Segments Tracked: %d", len(u.segmentSampleCounts))

   // Validate STSC sum
   var stscSum uint32
   for _, c := range u.segmentSampleCounts {
      stscSum += c
   }
   log.Printf("  Total Samples (stsc sum): %d", stscSum)
   log.Printf("  Total Samples (stsz len): %d", len(u.samples))
   log.Printf("  Total Duration: %d", totalDuration)

   if stscSum != uint32(len(u.samples)) {
      log.Printf("[Unfrag] ERROR: Sample count mismatch! stsc says %d, stsz says %d", stscSum, len(u.samples))
   } else {
      log.Println("[Unfrag] Sample counts match.")
   }
   // ---------------------

   // 3. Update Moov
   trak, _ := u.moov.Trak()
   mdia, _ := trak.Mdia()
   minf, _ := mdia.Minf()
   stbl, _ := minf.Stbl()

   mdhd, ok := mdia.Mdhd()
   if !ok {
      return errors.New("corrupt init segment: missing mdhd")
   }
   if err := patchDuration(mdhd.RawData, totalDuration); err != nil {
      return fmt.Errorf("patching mdhd: %w", err)
   }

   filterMvex(u.moov)

   var newChildren []StblChild
   if stsd, ok := stbl.Stsd(); ok {
      // Attempt to unprotect encryption here if needed
      stsd.UnprotectAll()
      newChildren = append(newChildren, StblChild{Stsd: stsd})
   } else {
      return errors.New("corrupt init segment: missing stsd")
   }

   newChildren = append(newChildren, StblChild{Raw: stts})
   newChildren = append(newChildren, StblChild{Raw: stsz})
   newChildren = append(newChildren, StblChild{Raw: stsc})
   newChildren = append(newChildren, StblChild{Raw: offsetBox})

   stbl.Children = newChildren

   // 4. Write Moov
   moovBytes := u.moov.Encode()
   if _, err := u.dst.Write(moovBytes); err != nil {
      return fmt.Errorf("writing moov: %w", err)
   }

   // 5. Update mdat Size
   if _, err := u.dst.Seek(u.mdatStartOffset+8, io.SeekStart); err != nil {
      return fmt.Errorf("seeking to mdat size: %w", err)
   }

   finalMdatSize := uint64(16) + u.payloadWritten
   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := u.dst.Write(sizeBuf[:]); err != nil {
      return fmt.Errorf("updating mdat size: %w", err)
   }

   u.dst.Seek(0, io.SeekEnd)
   log.Println("[Unfrag] Done.")
   return nil
}
