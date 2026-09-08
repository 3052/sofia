// remuxer.go
package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

// ErrStopped is returned by Process when the Stop channel was closed.
// Every segment processed before the return is durable in the output
// file; Progress reports where a resume should re-request from.
var ErrStopped = errors.New("process stopped")

// readBox buffers a complete box, header included, so the []byte
// decoders can work on it unchanged. Only for small boxes: never mdat.
func readBox(r io.Reader, typ [4]byte, size int64) ([]byte, error) {
   if size < 0 {
      return nil, errors.New("box extends to EOF; cannot buffer it")
   }
   if size > 0xFFFFFFFF {
      return nil, errors.New("box too large to buffer")
   }
   buf := make([]byte, size)
   binary.BigEndian.PutUint32(buf[0:4], uint32(size))
   copy(buf[4:8], typ[:])
   if _, err := io.ReadFull(r, buf[8:]); err != nil {
      return nil, err
   }
   return buf, nil
}

// readBoxHeader reads one box header from r and returns the box type
// with the full box size, header included. A size of -1 means the size
// field was 0 and the box extends to the end of the stream, which only
// makes sense for mdat.
func readBoxHeader(r io.Reader) (typ [4]byte, size int64, err error) {
   var head [8]byte
   if _, err = io.ReadFull(r, head[:]); err != nil {
      return
   }
   copy(typ[:], head[4:8])
   n := binary.BigEndian.Uint32(head[0:4])
   switch {
   case n == 0:
      size = -1
   case n == 1:
      var large [8]byte
      if _, err = io.ReadFull(r, large[:]); err != nil {
         return
      }
      size = int64(binary.BigEndian.Uint64(large[:]))
      if size < 16 {
         err = errors.New("invalid box largesize")
      }
   default:
      size = int64(n)
      if size < 8 {
         err = errors.New("invalid box size")
      }
   }
   return
}

// skipBox discards a box payload. size -1 means discard to EOF.
func skipBox(r io.Reader, size int64) error {
   if size < 0 {
      _, err := io.Copy(io.Discard, r)
      return err
   }
   _, err := io.CopyN(io.Discard, r, size)
   return err
}

type RemuxSample struct {
   Size                  uint32
   Duration              uint32
   IsSync                bool
   CompositionTimeOffset int32
}

type Remuxer struct {
   Writer   io.WriteSeeker
   Moov     *MoovBox
   OnSample func(data []byte, sample *SencSample)
   // OnMoov, when set, is called by Process when it consumes a moov from
   // the stream, before any segment is processed. Returning an error
   // aborts Process. This is where a caller fetches its DRM key: the KID
   // lives in the moov, and the key must be in hand before any sample is
   // decrypted. While the callback runs the reader is simply not being
   // read; the open HTTP connection sits backpressured.
   OnMoov func(moov *MoovBox) error
   // Stop, when closed, asks Process to return ErrStopped at the next
   // segment boundary. A nil Stop means Process never stops early. The
   // caller owns the channel and closes it; sofia only receives.
   Stop <-chan struct{}

   samples             []*RemuxSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   segmentCount        int

   // Process state: resumeOffset is the byte position, relative to the
   // first byte this session's Process call read, just past the last
   // complete segment.
   resumeOffset int64
}

// AddSegment processes one complete, standalone segment: every
// moof+mdat pair inside it becomes a fragment. Trailing bytes that do
// not form a complete box are ignored.
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
   stbl.RawChildren = append(stbl.RawChildren, stsz)
   stbl.RawChildren = append(stbl.RawChildren, stsc)
   stbl.RawChildren = append(stbl.RawChildren, offsetBox)
   if stss != nil {
      stbl.RawChildren = append(stbl.RawChildren, stss)
   }
   if ctts != nil {
      stbl.RawChildren = append(stbl.RawChildren, ctts)
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
   return r.beginMoovFrom(r.Moov)
}

// Process streams a continuous MP4 from reader: ftyp, sidx, styp and
// free boxes are skipped without interpretation, the first moov
// initializes the remuxer (calling OnMoov before any segment is
// processed; a moov arriving when the remuxer is already initialized —
// from an init segment or a resume — is skipped), and each moof+mdat
// pair becomes a segment processed exactly as AddSegment would process
// it. One segment at a time is held in memory — the moof and mdat
// bytes — and released before the next segment is read. The reader is
// consumed directly with no intermediate buffering: it should be the
// raw HTTP response body.
//
// It returns nil at EOF, ErrStopped when Stop was closed, or the first
// error. Stop is checked between segments, so a segment whose moof was
// already read still completes.
func (r *Remuxer) Process(reader io.Reader) error {
   br := &boxReader{R: reader}
   var pendingMoof *MoofBox
   for {
      if r.Stop != nil {
         select {
         case <-r.Stop:
            return ErrStopped
         default:
         }
      }
      typ, size, err := readBoxHeader(br)
      if err == io.EOF {
         return nil
      }
      if err != nil {
         return err
      }
      switch string(typ[:]) {
      case "moov":
         if r.Moov != nil {
            err = skipBox(br, size)
            break
         }
         data, bufErr := readBox(br, typ, size)
         if bufErr != nil {
            err = bufErr
            break
         }
         moov, decErr := DecodeMoovBox(data)
         if decErr != nil {
            err = decErr
            break
         }
         if err = r.beginMoovFrom(moov); err != nil {
            break
         }
         if r.OnMoov != nil {
            err = r.OnMoov(moov)
         }
      case "moof":
         data, bufErr := readBox(br, typ, size)
         if bufErr != nil {
            err = bufErr
            break
         }
         pendingMoof, err = DecodeMoofBox(data)
      case "mdat":
         if pendingMoof == nil {
            err = skipBox(br, size)
            break
         }
         var mdat *MdatBox
         if size < 0 {
            // box extends to EOF: the payload is the rest of the stream
            data, readErr := io.ReadAll(br)
            if readErr != nil {
               err = readErr
               break
            }
            mdat = &MdatBox{Payload: data}
         } else {
            data, bufErr := readBox(br, typ, size)
            if bufErr != nil {
               err = bufErr
               break
            }
            mdat, err = DecodeMdatBox(data)
            if err != nil {
               break
            }
         }
         if err = r.processFragment(pendingMoof, mdat); err != nil {
            break
         }
         pendingMoof = nil
         r.segmentCount++
         r.resumeOffset = br.N
      default:
         err = skipBox(br, size)
      }
      if err != nil {
         return err
      }
   }
}

// Progress reports the resume point after Process returns: the number
// of bytes consumed just past the last complete segment, and the number
// of segments processed. The offset is relative to the first byte this
// session's Process call read, so a resumed session adds the offset it
// started from.
func (r *Remuxer) Progress() (inputOffset int64, segmentsDone int) {
   return r.resumeOffset, r.segmentCount
}

// beginMoovFrom installs a moov decoded from a stream or an init segment
// and writes the placeholder mdat header at the current writer
// position. Initialize and Process both go through here.
func (r *Remuxer) beginMoovFrom(moov *MoovBox) error {
   if r.Writer == nil {
      return errors.New("writer is nil")
   }
   if len(moov.Trak) == 0 {
      return errors.New("no trak found")
   }
   r.Moov = moov
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

// boxReader counts every byte consumed so Progress can report a
// byte-exact resume offset.
type boxReader struct {
   R io.Reader
   N int64
}

func (b *boxReader) Read(p []byte) (int, error) {
   n, err := b.R.Read(p)
   b.N += int64(n)
   return n, err
}

// remuxer.go
