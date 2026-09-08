package sofia

import (
   "errors"
   "fmt"
   "io"
)

// ErrStopped is returned by Process when RequestStop was called. Every
// fragment fully processed before the return is durable in the output
// file; call Finish and Progress to save resume state.
var ErrStopped = errors.New("download stopped")

// RemuxSample is the remuxer's per-sample bookkeeping while fragments
// stream through; Finish turns the accumulated samples into the sample
// tables of a progressive MP4.
type RemuxSample struct {
   Size                  uint32
   Duration              uint32
   IsSync                bool
   CompositionTimeOffset int32
}

type Remuxer struct {
   // Writer is the output file. It must support Seek, and must not be
   // wrapped in anything that buffers writes across seeks.
   Writer io.WriteSeeker
   // Moov is the movie header, set by Initialize, AdoptState, or by
   // Process when it consumes a moov from the stream.
   Moov *MoovBox
   // OnSample, when set, is called with each sample's bytes just after
   // they are read and just before they are written. The slice points
   // into a reusable buffer that is only valid until the next sample;
   // decrypting in place is the intended use. Callers that retain the
   // slice must copy it.
   OnSample func(data []byte, sample *SencSample)
   // InitData holds the raw moov box bytes as consumed from the init
   // segment or from the head of a streamed file. Persist it next to a
   // stopped download and pass it to AdoptState to resume.
   InitData []byte

   samples             []*RemuxSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   segmentCount        int
   sampleBuf           []byte // reused per-sample read buffer, grown to max sample size
   input               *countingReader
   resumeOffset        int64
   stop                chan struct{}
}

// NewRemuxer creates a remuxer writing to w. Prefer this over a zero
// value: it pre-creates the stop channel, so RequestStop is safe to call
// from another goroutine at any time.
func NewRemuxer(w io.WriteSeeker) *Remuxer {
   return &Remuxer{Writer: w, stop: make(chan struct{})}
}

// Process streams a complete MP4: any sequence of ftyp, moov, sidx, styp,
// moof and mdat boxes. If a moov has not been installed yet — via
// Initialize or AdoptState — it is consumed from the stream itself, so a
// SegmentBase file can be fed in with a single request. Control boxes are
// buffered (they are small); mdat payloads are processed sample by sample
// through a reusable buffer, so peak memory for media data is one sample.
//
// Feeding one downloaded segment at a time also works:
//
// Process(bytes.NewReader(segmentData))
//
// The reader must not be wrapped in bufio; the byte counting that backs
// Progress would run ahead of actual consumption.
func (r *Remuxer) Process(reader io.Reader) error {
   if r.input != nil {
      return errors.New("Process already running")
   }
   if r.stop == nil {
      r.stop = make(chan struct{})
   }
   r.input = &countingReader{R: reader}
   defer func() { r.input = nil }()

   var pendingMoof *MoofBox
   for {
      select {
      case <-r.stop:
         return ErrStopped
      default:
      }
      typ, payloadLen, err := readBoxHeader(r.input)
      if err == io.EOF {
         return nil
      }
      if err != nil {
         return err
      }
      switch string(typ[:]) {
      case "moov":
         if r.Moov != nil {
            // in-band moov; we already have one
            if err := drain(r.input, payloadLen); err != nil {
               return err
            }
            continue
         }
         data, err := readWholeBox(r.input, typ, payloadLen)
         if err != nil {
            return err
         }
         moov, err := DecodeMoovBox(data)
         if err != nil {
            return fmt.Errorf("parsing moov: %w", err)
         }
         if err := r.startOutput(moov, data); err != nil {
            return err
         }
      case "moof":
         if r.Moov == nil {
            return errors.New("moof before moov: call Initialize first or stream from the start")
         }
         data, err := readWholeBox(r.input, typ, payloadLen)
         if err != nil {
            return err
         }
         moof, err := DecodeMoofBox(data)
         if err != nil {
            return fmt.Errorf("parsing moof: %w", err)
         }
         pendingMoof = moof
      case "mdat":
         if pendingMoof == nil {
            if err := drain(r.input, payloadLen); err != nil {
               return err
            }
            continue
         }
         // If stop was requested while the moof was being read, do not
         // start its mdat: resumeOffset still points before the moof, so
         // the pair is simply re-fetched on resume.
         select {
         case <-r.stop:
            return ErrStopped
         default:
         }
         if err := r.processFragmentReader(pendingMoof, r.input, payloadLen); err != nil {
            return err
         }
         pendingMoof = nil
         r.segmentCount++
         r.resumeOffset = r.input.N
      default:
         // ftyp, sidx, styp, free, emsg, prft: skipped without interpretation
         if err := drain(r.input, payloadLen); err != nil {
            return err
         }
      }
   }
}

// Progress returns the resume point for a single-URL stream: the byte
// offset into the input just past the last fully processed fragment, and
// the number of fragments completed. Pass the offset as the start of a
// Range header to resume. Only meaningful when the whole file was fed to
// one Process call; call it after Process returns.
func (r *Remuxer) Progress() (inputOffset int64, segmentsDone int) {
   return r.resumeOffset, r.segmentCount
}

// RequestStop asks Process to return after the fragment currently being
// processed completes. Safe to call from another goroutine, and safe to
// call more than once. If the remuxer was not created with NewRemuxer and
// Process has not started yet, this is a no-op.
func (r *Remuxer) RequestStop() {
   if r.stop == nil {
      return
   }
   select {
   case <-r.stop:
   default:
      close(r.stop)
   }
}

// processFragmentReader streams one moof's mdat payload: sample bytes are
// read into the reusable buffer (at most one sample in memory at a time),
// passed through OnSample, and written to the output as one chunk. The
// trun data offset is ignored: samples are assumed to start at the start
// of the mdat payload, as fragment writers in practice do.
func (r *Remuxer) processFragmentReader(moof *MoofBox, reader io.Reader, payloadLen int64) error {
   traf := moof.Traf
   if traf == nil || traf.Tfhd == nil {
      return drain(reader, payloadLen)
   }
   senc := traf.Senc
   sencIndex := 0
   defDur := traf.Tfhd.DefaultSampleDuration
   defSize := traf.Tfhd.DefaultSampleSize
   defFlags := traf.Tfhd.DefaultSampleFlags
   remaining := payloadLen // -1 = unbounded (size-0 mdat)

   readSample := func(n uint32) error {
      if remaining >= 0 && int64(n) > remaining {
         return errors.New("mdat payload too short for samples")
      }
      if uint32(cap(r.sampleBuf)) < n {
         r.sampleBuf = make([]byte, n)
      }
      buf := r.sampleBuf[:n]
      if _, err := io.ReadFull(reader, buf); err != nil {
         return err
      }
      if remaining >= 0 {
         remaining -= int64(n)
      }
      return nil
   }

   var newSamples []*RemuxSample
   wroteChunk := false
   for _, trun := range traf.Trun {
      for i, sample := range trun.Samples {
         remuxSample := &RemuxSample{
            Duration:              defDur,
            Size:                  defSize,
            IsSync:                true,
            CompositionTimeOffset: 0,
         }
         currentFlags := defFlags
         // NOTE: The order of these two flag checks matters!
         // Per ISO/IEC 14496-12, if both sample_flags_present (0x000400) and
         // first_sample_flags_present (0x000004) are set, FirstSampleFlags
         // must OVERRIDE sample.Flags for the first sample (i==0).
         // Therefore, we must check sample_flags_present FIRST, then let
         // first_sample_flags_present overwrite it for i==0.
         // DO NOT swap these blocks, or FirstSampleFlags will be clobbered
         // by sample.Flags and the keyframe (sync sample) detection will be
         // corrupted for the first sample of each trun.
         if (trun.Flags & 0x000400) != 0 {
            currentFlags = sample.Flags
         }
         if i == 0 && (trun.Flags&0x000004) != 0 {
            currentFlags = trun.FirstSampleFlags
         }
         if (trun.Flags & 0x000100) != 0 {
            remuxSample.Duration = sample.Duration
         }
         if (trun.Flags & 0x000200) != 0 {
            remuxSample.Size = sample.Size
         }
         if (trun.Flags & 0x000800) != 0 {
            remuxSample.CompositionTimeOffset = sample.CompositionTimeOffset
         }
         if (currentFlags & 0x00010000) != 0 {
            remuxSample.IsSync = false
         } else {
            remuxSample.IsSync = true
         }
         if err := readSample(remuxSample.Size); err != nil {
            return err
         }
         sampleData := r.sampleBuf[:remuxSample.Size]
         var encInfo *SencSample
         if senc != nil && sencIndex < len(senc.Samples) {
            encInfo = &senc.Samples[sencIndex]
            sencIndex++
         }
         if r.OnSample != nil {
            r.OnSample(sampleData, encInfo)
         }
         if !wroteChunk {
            currentPos, err := r.Writer.Seek(0, io.SeekCurrent)
            if err != nil {
               return fmt.Errorf("seeking to get chunk offset: %w", err)
            }
            r.chunkOffsets = append(r.chunkOffsets, uint64(currentPos))
            wroteChunk = true
         }
         if _, err := r.Writer.Write(sampleData); err != nil {
            return err
         }
         newSamples = append(newSamples, remuxSample)
      }
   }

   if len(newSamples) == 0 {
      return drain(reader, payloadLen)
   }
   r.samples = append(r.samples, newSamples...)
   r.segmentSampleCounts = append(r.segmentSampleCounts, uint32(len(newSamples)))
   // drain any bytes the sample sizes did not cover
   if remaining > 0 {
      return drain(reader, remaining)
   }
   return nil
}

// remuxer.go
