package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

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

   // Feed state: buf holds the bytes of a box that is not yet complete,
   // bytesFed counts every byte ever fed, pendingMoof is the moof whose
   // mdat has not fully arrived, and resumeOffset is the input position
   // (relative to the first byte fed) just past the last complete
   // fragment.
   buf          []byte
   bytesFed     int64
   pendingMoof  *MoofBox
   resumeOffset int64
}

// AddSegment processes one complete, standalone segment: every
// moof+mdat pair inside it becomes a fragment. Trailing bytes that do
// not form a complete box are ignored. Callers feeding a continuous
// single-URL stream must use Feed instead, which keeps incomplete boxes
// buffered across calls.
func (r *Remuxer) AddSegment(segmentData []byte) error {
   if r.Moov == nil {
      return errors.New("must call Initialize")
   }
   r.segmentCount++
   boxes, err := DecodeBoxes(segmentData)
   if err != nil {
      return fmt.Errorf("parsing segment %d: %w", r.segmentCount, err)
   }
   var pendingMoof *MoofBox
   for i, box := range boxes {
      if box.Moof != nil {
         pendingMoof = box.Moof
         continue
      }
      if box.Mdat != nil {
         if pendingMoof != nil {
            if err := r.processFragment(pendingMoof, box.Mdat); err != nil {
               return fmt.Errorf("processing fragment at box index %d: %w", i, err)
            }
            pendingMoof = nil
         }
      }
   }
   return nil
}

// AdoptState resumes a remuxer from state recovered out of a stopped file
// (StateFromMoov). The mdat header written by Initialize sits at offset 0
// of the original file, where the payloads still begin, so the chunk
// offsets in the state remain valid as-is. The caller truncates the old
// moov away and positions the writer at the payload boundary; AdoptState
// itself performs no I/O.
func (r *Remuxer) AdoptState(initSegment []byte, state *RemuxState, segmentsDone int) error {
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   if state == nil {
      return errors.New("state is nil")
   }
   if len(state.ChunkOffsets) != len(state.SamplesPerChunk) {
      return errors.New("inconsistent resume state")
   }
   if err := r.initMoov(initSegment); err != nil {
      return err
   }
   r.mdatStartOffset = 0
   r.samples = state.Samples
   r.chunkOffsets = state.ChunkOffsets
   r.segmentSampleCounts = state.SamplesPerChunk
   r.segmentCount = segmentsDone
   return nil
}

// Feed streams a continuous single-URL file in slices: data is appended
// to an internal buffer and every box that is fully buffered is
// processed — ftyp, sidx and styp are skipped, and each complete
// moof+mdat pair becomes a fragment. Bytes of a box that is not yet
// complete stay buffered until a later Feed completes it, so the buffer
// never grows past the largest incomplete box, which in practice is the
// mdat of one fragment. Initialize must be called first.
func (r *Remuxer) Feed(data []byte) error {
   if r.Moov == nil {
      return errors.New("must call Initialize")
   }
   r.buf = append(r.buf, data...)
   r.bytesFed += int64(len(data))
   for len(r.buf) >= 8 {
      header, err := DecodeBoxHeader(r.buf)
      if err != nil {
         return err
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         // box extends to the end of the data
         boxSize = len(r.buf)
      }
      if boxSize < 8 {
         return errors.New("invalid box size")
      }
      if boxSize > len(r.buf) {
         // incomplete box: wait for more data
         return nil
      }
      boxData := r.buf[:boxSize]
      switch string(header.Type[:]) {
      case "moof":
         moof, err := DecodeMoofBox(boxData)
         if err != nil {
            return fmt.Errorf("parsing moof: %w", err)
         }
         r.pendingMoof = moof
      case "mdat":
         if r.pendingMoof != nil {
            mdat, err := DecodeMdatBox(boxData)
            if err != nil {
               return err
            }
            if err := r.processFragment(r.pendingMoof, mdat); err != nil {
               return err
            }
            r.pendingMoof = nil
            r.segmentCount++
            // input position (relative to the first byte fed) just past
            // this fragment
            r.resumeOffset = r.bytesFed - int64(len(r.buf)-boxSize)
         }
      default:
         // ftyp, sidx, styp, moov (already installed): skipped
      }
      r.buf = r.buf[boxSize:]
   }
   return nil
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
   if err := r.initMoov(initSegment); err != nil {
      return err
   }
   var err error
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

// Progress returns the resume point for a single-URL stream: the input
// byte offset just past the last fully processed fragment — relative to
// the FIRST byte fed to this Remuxer, so a resumed session must add its
// own starting offset — and the number of fragments completed. Bytes of
// a fragment that was still incomplete when the download stopped are
// re-fed on resume.
func (r *Remuxer) Progress() (inputOffset int64, fragmentsDone int) {
   return r.resumeOffset, r.segmentCount
}

func (r *Remuxer) initMoov(initSegment []byte) error {
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
   return nil
}

// remuxer.go
