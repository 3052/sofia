package sofia

import (
   "bytes"
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

// --- MOOV ---
type MoovBox struct {
   Header      *BoxHeader
   Mvhd        *MvhdBox
   Trak        []*TrakBox
   Pssh        []*PsshBox
   RawChildren [][]byte
}

func DecodeMoovBox(data []byte) (*MoovBox, error) {
   b := &MoovBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   payload := data[8:b.Header.Size]
   offset := 0
   for offset < len(payload) {
      header, err := DecodeBoxHeader(payload[offset:])
      if err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(payload) - offset
      }
      if boxSize < 8 || offset+boxSize > len(payload) {
         return nil, errors.New("invalid child box size")
      }

      content := payload[offset : offset+boxSize]
      switch string(header.Type[:]) {
      case "mvhd":
         mvhd, err := DecodeMvhdBox(content)
         if err != nil {
            return nil, err
         }
         b.Mvhd = mvhd
      case "trak":
         trak, err := DecodeTrakBox(content)
         if err != nil {
            return nil, err
         }
         b.Trak = append(b.Trak, trak)
      case "pssh":
         pssh, err := DecodePsshBox(content)
         if err != nil {
            return nil, err
         }
         b.Pssh = append(b.Pssh, pssh)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      offset += boxSize
   }
   return b, nil
}

// extractMoov scans a byte slice's top-level boxes for the first moov and
// decodes it, also returning the moov's raw bytes. Used on init segments
// and on the raw bytes kept in InitData.
func extractMoov(data []byte) (*MoovBox, []byte, error) {
   offset := 0
   for offset+8 <= len(data) {
      header, err := DecodeBoxHeader(data[offset:])
      if err != nil {
         break
      }
      size := int(header.Size)
      if size == 0 {
         size = len(data) - offset
      }
      if size < 8 || offset+size > len(data) {
         return nil, nil, errors.New("invalid box size in init segment")
      }
      if string(header.Type[:]) == "moov" {
         raw := data[offset : offset+size]
         moov, err := DecodeMoovBox(raw)
         if err != nil {
            return nil, nil, err
         }
         return moov, raw, nil
      }
      offset += size
   }
   return nil, nil, errors.New("no moov found")
}

func (b *MoovBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Mvhd != nil {
      buffer = append(buffer, b.Mvhd.Encode()...)
   }
   for _, trak := range b.Trak {
      buffer = append(buffer, trak.Encode()...)
   }
   // pssh is skipped on encode
   for _, raw := range b.RawChildren {
      buffer = append(buffer, raw...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// FindDefaultKID walks the first track's sample entries and returns the first
// non-zero default KID declared by a 'tenc' box, following
// trak > mdia > minf > stbl > stsd > enc > sinf > schi > tenc.
// An all-zero KID means no key is declared by that box, so it is skipped.
// Returns nil if no KID is found.
func (b *MoovBox) FindDefaultKID() []byte {
   if len(b.Trak) == 0 {
      return nil
   }
   trak := b.Trak[0]
   if trak.Mdia == nil || trak.Mdia.Minf == nil || trak.Mdia.Minf.Stbl == nil || trak.Mdia.Minf.Stbl.Stsd == nil {
      return nil
   }
   for _, enc := range trak.Mdia.Minf.Stbl.Stsd.EncChildren {
      if enc.Sinf == nil || enc.Sinf.Schi == nil || enc.Sinf.Schi.Tenc == nil {
         continue
      }
      kid := enc.Sinf.Schi.Tenc.DefaultKID
      if kid != ([16]byte{}) {
         return kid[:]
      }
   }
   return nil
}

func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, pssh := range b.Pssh {
      if bytes.Equal(pssh.SystemID[:], systemID) {
         return pssh, true
      }
   }
   return nil, false
}

func (b *MoovBox) RemoveMvex() {
   var kept [][]byte
   for _, child := range b.RawChildren {
      if len(child) >= 8 && string(child[4:8]) == "mvex" {
         continue
      }
      kept = append(kept, child)
   }
   b.RawChildren = kept
}

// --- MVHD ---
type MvhdBox struct {
   Header           *BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   RemainingData    []byte
}

func DecodeMvhdBox(data []byte) (*MvhdBox, error) {
   b := &MvhdBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 12 {
      return nil, errors.New("mvhd box too small")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Bytes(4)
   b.Version = versionAndFlags[0]
   copy(b.Flags[:], versionAndFlags[1:])

   if b.Version == 1 {
      if len(data) < 40 { // 8 header + 4 version/flags + 28 v1 body
         return nil, errors.New("mvhd v1 too short")
      }
      b.CreationTime = p.Uint64()
      b.ModificationTime = p.Uint64()
      b.Timescale = p.Uint32()
      b.Duration = p.Uint64()
   } else { // Version 0
      if len(data) < 28 { // 8 header + 4 version/flags + 16 v0 body
         return nil, errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(p.Uint32())
      b.ModificationTime = uint64(p.Uint32())
      b.Timescale = p.Uint32()
      b.Duration = uint64(p.Uint32())
   }

   b.RemainingData = data[p.offset:b.Header.Size]
   return b, nil
}

func (b *MvhdBox) Encode() []byte {
   var bodySize int
   if b.Version == 1 {
      bodySize = 32 // 8+8+4+8 + 4 for ver/flags
   } else {
      bodySize = 20 // 4+4+4+4 + 4 for ver/flags
   }
   totalSize := uint32(8 + bodySize + len(b.RemainingData))
   buffer := make([]byte, totalSize)

   w := writer{buf: buffer}
   w.PutUint32(totalSize)
   w.PutBytes(b.Header.Type[:])
   w.PutByte(b.Version)
   w.PutBytes(b.Flags[:])

   if b.Version == 1 {
      w.PutUint64(b.CreationTime)
      w.PutUint64(b.ModificationTime)
      w.PutUint32(b.Timescale)
      w.PutUint64(b.Duration)
   } else {
      w.PutUint32(uint32(b.CreationTime))
      w.PutUint32(uint32(b.ModificationTime))
      w.PutUint32(b.Timescale)
      w.PutUint32(uint32(b.Duration))
   }

   w.PutBytes(b.RemainingData)
   b.Header.Size = totalSize
   return buffer
}

func (b *MvhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

// Initialize installs the moov from a downloaded init segment (ftyp,
// moov, ...) and writes the placeholder mdat header at the current writer
// position. Keep the init segment bytes (or take them from InitData
// afterwards) to pass to AdoptState when resuming.
func (r *Remuxer) Initialize(initSegment []byte) error {
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   moov, raw, err := extractMoov(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init segment: %w", err)
   }
   return r.startOutput(moov, raw)
}

// startOutput installs moov and writes the placeholder mdat header at the
// current writer position. Initialize and Process both go through here
// after extracting a moov.
func (r *Remuxer) startOutput(moov *MoovBox, raw []byte) error {
   if r.Writer == nil {
      return errors.New("writer is nil")
   }
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   if len(moov.Trak) == 0 {
      return errors.New("no trak found")
   }
   r.Moov = moov
   r.InitData = raw
   pos, err := r.Writer.Seek(0, io.SeekCurrent)
   if err != nil {
      return fmt.Errorf("seeking to get current position: %w", err)
   }
   r.mdatStartOffset = pos
   header := make([]byte, 16)
   binary.BigEndian.PutUint32(header[0:4], 1)
   copy(header[4:8], []byte("mdat"))
   _, err = r.Writer.Write(header)
   return err
}

// AdoptState resumes a remuxer from state recovered out of a stopped file
// (StateFromMoov). The mdat header written by the first run sits at offset
// 0 of the output file, where the payloads still begin, so the chunk
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
   moov, raw, err := extractMoov(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init segment: %w", err)
   }
   if r.Writer == nil {
      return errors.New("writer is nil")
   }
   if len(moov.Trak) == 0 {
      return errors.New("no trak found")
   }
   r.Moov = moov
   r.InitData = raw
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

// movie.go
