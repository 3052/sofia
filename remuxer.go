package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

// AppendResult describes the fragments an AddSegment call appended to the file.
type AppendResult struct {
   EndOffset       int64
   ChunkOffsets    []uint64
   SamplesPerChunk []uint32
   Samples         []*RemuxSample
}

type RemuxSample struct {
   Size                  uint32
   Duration              uint32
   IsSync                bool
   CompositionTimeOffset int32
}

type Remuxer struct {
   Writer              io.WriteSeeker
   Moov                *MoovBox
   samples             []*RemuxSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   segmentCount        int
   OnSample            func(data []byte, sample *SencSample)
}

func (r *Remuxer) AddSegment(segmentData []byte) (*AppendResult, error) {
   if r.Moov == nil {
      return nil, errors.New("must call Initialize")
   }
   r.segmentCount++
   boxes, err := DecodeBoxes(segmentData)
   if err != nil {
      return nil, fmt.Errorf("parsing segment %d: %w", r.segmentCount, err)
   }
   sampleBase := len(r.samples)
   chunkBase := len(r.chunkOffsets)
   var pendingMoof *MoofBox
   for i, box := range boxes {
      if box.Moof != nil {
         pendingMoof = box.Moof
         continue
      }
      if box.Mdat != nil {
         if pendingMoof != nil {
            if err := r.processFragment(pendingMoof, box.Mdat); err != nil {
               return nil, fmt.Errorf("processing fragment at box index %d: %w", i, err)
            }
            pendingMoof = nil
         }
      }
   }
   endOffset, err := r.Writer.Seek(0, io.SeekCurrent)
   if err != nil {
      return nil, fmt.Errorf("seeking to get segment end offset: %w", err)
   }
   return &AppendResult{
      EndOffset:       endOffset,
      ChunkOffsets:    r.chunkOffsets[chunkBase:],
      SamplesPerChunk: r.segmentSampleCounts[chunkBase:],
      Samples:         r.samples[sampleBase:],
   }, nil
}

func (r *Remuxer) Finish() error {
   if r.Moov == nil {
      return errors.New("not initialized")
   }
   mdatEndOffset, err := r.Writer.Seek(0, io.SeekCurrent)
   if err != nil {
      return fmt.Errorf("seeking to get mdat end offset: %w", err)
   }
   finalMdatSize := uint64(mdatEndOffset - r.mdatStartOffset)
   var totalDuration uint64
   for _, sample := range r.samples {
      totalDuration += uint64(sample.Duration)
   }
   stts := buildStts(r.samples)
   stsz := buildStsz(r.samples)
   stsc := buildStsc(r.segmentSampleCounts)
   offsetBox := buildChunkOffsetBox(r.chunkOffsets)
   stss := buildStss(r.samples)
   ctts := buildCtts(r.samples)

   if len(r.Moov.Trak) == 0 {
      return errors.New("cannot finish remux: no trak in moov")
   }
   trak := r.Moov.Trak[0]
   if trak.Mdia == nil {
      return errors.New("missing mdia")
   }
   mdia := trak.Mdia
   if mdia.Minf == nil {
      return errors.New("missing minf")
   }
   minf := mdia.Minf
   if minf.Stbl == nil {
      return errors.New("missing stbl")
   }
   stbl := minf.Stbl
   mdhd := mdia.Mdhd
   if mdhd == nil {
      return errors.New("missing mdhd")
   }
   mdhd.SetDuration(totalDuration)
   if mvhd := r.Moov.Mvhd; mvhd != nil {
      mvhd.Timescale = mdhd.Timescale
      mvhd.SetDuration(totalDuration)
   }
   r.Moov.RemoveMvex()
   trak.RemoveEdts()
   stbl.RawChildren = nil // Clear existing table boxes
   if stbl.Stsd == nil {
      return errors.New("missing stsd")
   }
   stbl.Stsd.RemoveSinf()
   stbl.RawChildren = append(stbl.RawChildren, stts)
   if ctts != nil {
      stbl.RawChildren = append(stbl.RawChildren, ctts)
   }
   stbl.RawChildren = append(stbl.RawChildren, stsz)
   stbl.RawChildren = append(stbl.RawChildren, stsc)
   stbl.RawChildren = append(stbl.RawChildren, offsetBox)
   if stss != nil {
      stbl.RawChildren = append(stbl.RawChildren, stss)
   }
   moovBytes := r.Moov.Encode()
   if _, err := r.Writer.Write(moovBytes); err != nil {
      return err
   }
   if _, err := r.Writer.Seek(r.mdatStartOffset+8, io.SeekStart); err != nil {
      return fmt.Errorf("seeking to patch mdat size: %w", err)
   }
   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := r.Writer.Write(sizeBuf[:]); err != nil {
      return err
   }
   if _, err := r.Writer.Seek(0, io.SeekEnd); err != nil {
      return fmt.Errorf("seeking to end of file: %w", err)
   }
   return nil
}

func (r *Remuxer) Initialize(initSegment []byte) error {
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   if r.Writer == nil {
      return errors.New("writer is nil")
   }
   boxes, err := DecodeBoxes(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init segment: %w", err)
   }
   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found")
   }
   r.Moov = moovPtr
   if len(r.Moov.Trak) == 0 {
      return errors.New("no trak found")
   }
   r.mdatStartOffset, err = r.Writer.Seek(0, io.SeekCurrent)
   if err != nil {
      return fmt.Errorf("seeking to get current position: %w", err)
   }
   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))
   _, err = r.Writer.Write(mdatHeader)
   return err
}

// Resume prepares the remuxer to continue writing a file that was
// interrupted mid-download. The file must have been created by Initialize,
// so its extended-size 'mdat' box sits at the start of the file. samples,
// chunkOffsets and samplesPerChunk must describe the fragments already
// present in the file, and segmentsDone is the number of AddSegment calls
// that produced them. On success the writer is positioned at the end of
// the valid data, ready for AddSegment.
func (r *Remuxer) Resume(initSegment []byte, segmentsDone int, samples []*RemuxSample, chunkOffsets []uint64, samplesPerChunk []uint32) error {
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   if r.Writer == nil {
      return errors.New("writer is nil")
   }
   if len(chunkOffsets) != len(samplesPerChunk) {
      return errors.New("inconsistent resume state")
   }
   boxes, err := DecodeBoxes(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init segment: %w", err)
   }
   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found")
   }
   r.Moov = moovPtr
   if len(r.Moov.Trak) == 0 {
      return errors.New("no trak found")
   }
   // Initialize always writes the mdat header at the start of a fresh file.
   r.mdatStartOffset = 0
   r.samples = samples
   r.chunkOffsets = chunkOffsets
   r.segmentSampleCounts = samplesPerChunk
   r.segmentCount = segmentsDone
   if _, err := r.Writer.Seek(0, io.SeekEnd); err != nil {
      return fmt.Errorf("seeking to end of file: %w", err)
   }
   return nil
}

// remuxer.go
