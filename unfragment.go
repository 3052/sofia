package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

// Unfragmenter handles the conversion of fragmented segments into a single file
// statefully. It keeps memory usage low by streaming payloads immediately.
type Unfragmenter struct {
   dst  io.WriteSeeker
   moov *MoovBox // The template moov from init segment
   ftyp []byte

   // Track state
   samples             []sampleInfo
   chunkOffsets        []uint64 // Absolute file offsets of each segment's payload
   segmentSampleCounts []uint32 // Needed for stsc

   // Byte tracking
   mdatStartOffset int64  // Where the mdat box starts in the file
   payloadWritten  uint64 // Total bytes of media data written so far

   initialized bool
}

type sampleInfo struct {
   Size              uint32
   Duration          uint32
   IsSync            bool
   CompositionOffset int32
}

// NewUnfragmenter creates a new converter writing to the provided file.
func NewUnfragmenter(dst io.WriteSeeker) *Unfragmenter {
   return &Unfragmenter{dst: dst}
}

// Initialize processes the initialization segment (init.mp4).
// It writes the ftyp box and a placeholder mdat header to the output.
func (u *Unfragmenter) Initialize(initSegment []byte) error {
   if u.initialized {
      return errors.New("already initialized")
   }

   boxes, err := Parse(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init: %w", err)
   }

   // 1. Capture ftyp
   for _, b := range boxes {
      // Fix: Removed 'b.Raw != nil' check (S1009)
      if len(b.Raw) >= 8 && string(b.Raw[4:8]) == "ftyp" {
         u.ftyp = b.Raw
         break
      }
   }

   // 2. Capture and validate Moov
   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found in init segment")
   }
   u.moov = moovPtr

   // Validate hierarchy
   if _, ok := u.moov.Trak(); !ok {
      return errors.New("no trak in moov")
   }

   // 3. Write ftyp to output
   if u.ftyp != nil {
      if _, err := u.dst.Write(u.ftyp); err != nil {
         return fmt.Errorf("writing ftyp: %w", err)
      }
   }

   // 4. Write mdat Header Placeholder
   // We assume 64-bit size (16 bytes) to be safe for multi-GB files.
   // [size: 1 (4b)] [type: mdat (4b)] [largeSize: 0 (8b placeholder)]
   u.mdatStartOffset, _ = u.dst.Seek(0, io.SeekCurrent)
   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))
   if _, err := u.dst.Write(mdatHeader); err != nil {
      return fmt.Errorf("writing mdat header: %w", err)
   }

   u.initialized = true
   return nil
}

// AddSegment processes a single media segment (e.g., segment-1.m4s).
// It parses metadata into memory and appends the payload immediately to disk.
func (u *Unfragmenter) AddSegment(segmentData []byte) error {
   if !u.initialized {
      return errors.New("must call Initialize before AddSegment")
   }

   boxes, err := Parse(segmentData)
   if err != nil {
      return fmt.Errorf("parsing segment: %w", err)
   }

   moof := FindMoofPtr(boxes)
   mdat := FindMdatPtr(boxes)

   if moof == nil || mdat == nil {
      // Possibly a silent audio segment or metadata only?
      // We skip it to maintain stream integrity.
      return nil
   }

   // 1. Calculate absolute offset for this chunk
   // Current File Position is end of previous write.
   // But we need to account for the fact that we are inside an 'mdat' box logically.
   // The STCO/CO64 offset is absolute from start of file.
   currentPos, _ := u.dst.Seek(0, io.SeekCurrent)
   u.chunkOffsets = append(u.chunkOffsets, uint64(currentPos))

   // 2. Write Payload immediately
   n, err := u.dst.Write(mdat.Payload)
   if err != nil {
      return fmt.Errorf("writing payload: %w", err)
   }
   u.payloadWritten += uint64(n)

   // 3. Extract Metadata
   traf, ok := moof.Traf()
   if !ok {
      return nil
   }
   tfhd := traf.Tfhd()
   trun := traf.Trun()

   if tfhd != nil && trun != nil {
      // Defaults
      defDur, defSize, defFlags := tfhd.DefaultSampleDuration, tfhd.DefaultSampleSize, tfhd.DefaultSampleFlags

      // Record sample count for this chunk (for stsc)
      u.segmentSampleCounts = append(u.segmentSampleCounts, uint32(len(trun.Samples)))

      for _, s := range trun.Samples {
         si := sampleInfo{
            Duration: defDur,
            Size:     defSize,
            IsSync:   (defFlags & 0x00010000) == 0,
         }

         // Overrides
         if (trun.Flags & 0x000100) != 0 {
            si.Duration = s.Duration
         }
         if (trun.Flags & 0x000200) != 0 {
            si.Size = s.Size
         }
         if (trun.Flags & 0x000400) != 0 {
            si.IsSync = (s.Flags & 0x00010000) == 0
         }
         if (trun.Flags & 0x000800) != 0 {
            si.CompositionOffset = s.CompositionTimeOffset
         }

         u.samples = append(u.samples, si)
      }
   }

   return nil
}

// Finish constructs the final atoms, writes the moov box at the end of the file,
// and updates the mdat header size.
func (u *Unfragmenter) Finish() error {
   if !u.initialized {
      return errors.New("not initialized")
   }

   // 1. Build Tables
   stts := buildStts(u.samples)
   stsz := buildStsz(u.samples)
   stsc := buildStsc(u.segmentSampleCounts)
   stss := buildStss(u.samples)
   ctts := buildCtts(u.samples)

   // 2. Build Offsets (co64 or stco)
   // Since we are writing moov at the end, the offsets we calculated in AddSegment
   // are already correct absolute offsets.
   var offsetBox []byte
   is64Bit := false

   // Check if any offset exceeds 4GB
   if len(u.chunkOffsets) > 0 {
      lastOffset := u.chunkOffsets[len(u.chunkOffsets)-1]
      if lastOffset > 0xFFFFFFFF {
         is64Bit = true
      }
   }

   if is64Bit {
      offsetBox = buildCo64(u.chunkOffsets)
   } else {
      offsetBox = buildStco(u.chunkOffsets)
   }

   // 3. Inject into Moov
   trak, _ := u.moov.Trak()
   mdia, _ := trak.Mdia()
   minf, _ := mdia.Minf()
   stbl, _ := minf.Stbl()

   // Filter mvex (removes fragmentation flag)
   filterMvex(u.moov)

   // Rebuild stbl children
   var newChildren []StblChild
   if stsd, ok := stbl.Stsd(); ok {
      newChildren = append(newChildren, StblChild{Stsd: stsd})
   } else {
      return errors.New("corrupt init segment: missing stsd")
   }

   newChildren = append(newChildren, StblChild{Raw: stts})
   newChildren = append(newChildren, StblChild{Raw: stsz})
   newChildren = append(newChildren, StblChild{Raw: stsc})
   newChildren = append(newChildren, StblChild{Raw: offsetBox}) // The calculated offsets
   if stss != nil {
      newChildren = append(newChildren, StblChild{Raw: stss})
   }
   if ctts != nil {
      newChildren = append(newChildren, StblChild{Raw: ctts})
   }

   stbl.Children = newChildren

   // 4. Write Moov to end of file
   moovBytes := u.moov.Encode()
   if _, err := u.dst.Write(moovBytes); err != nil {
      return fmt.Errorf("writing moov: %w", err)
   }

   // 5. Update mdat Size
   // Go back to the start of mdat header
   if _, err := u.dst.Seek(u.mdatStartOffset+8, io.SeekStart); err != nil {
      return fmt.Errorf("seeking to mdat size: %w", err)
   }

   // We used 16-byte header (Large Size). The size field is at offset 8.
   // Value = Header(16) + Payload
   finalMdatSize := uint64(16) + u.payloadWritten

   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := u.dst.Write(sizeBuf[:]); err != nil {
      return fmt.Errorf("updating mdat size: %w", err)
   }

   // Seek back to end just in case user does more writes (optional)
   u.dst.Seek(0, io.SeekEnd)

   return nil
}
