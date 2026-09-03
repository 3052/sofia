package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

// remuxer.go
func FindMoov(boxes []Box) (*MoovBox, bool) {
   for _, box := range boxes {
      if box.Moov != nil {
         return box.Moov, true
      }
   }
   return nil, false
}

// --- Box ---
type Box struct {
   Moov *MoovBox
   Moof *MoofBox
   Mdat *MdatBox
   Sidx *SidxBox
   Raw  []byte
}

func DecodeBoxes(data []byte) ([]Box, error) {
   var boxes []Box
   offset := 0
   for offset < len(data) {
      header, err := DecodeBoxHeader(data[offset:])
      if err != nil {
         break
      }
      boxSize := int(header.Size)
      if boxSize == 0 {
         boxSize = len(data) - offset
      }
      if boxSize < 8 || offset+boxSize > len(data) {
         return nil, errors.New("invalid child box size")
      }

      boxData := data[offset : offset+boxSize]
      var currentBox Box
      switch string(header.Type[:]) {
      case "moov":
         moov, err := DecodeMoovBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moov = moov
      case "moof":
         moof, err := DecodeMoofBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Moof = moof
      case "mdat":
         mdat, err := DecodeMdatBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Mdat = mdat
      case "sidx":
         sidx, err := DecodeSidxBox(boxData)
         if err != nil {
            return nil, err
         }
         currentBox.Sidx = sidx
      default:
         currentBox.Raw = boxData
      }
      boxes = append(boxes, currentBox)
      offset += boxSize
   }
   return boxes, nil
}

// --- BoxHeader ---
type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func DecodeBoxHeader(data []byte) (*BoxHeader, error) {
   if len(data) < 8 {
      return nil, errors.New("not enough data for box header")
   }
   h := &BoxHeader{}
   p := parser{data: data}
   h.Size = p.Uint32()
   copy(h.Type[:], p.Bytes(4))
   return h, nil
}

func (h *BoxHeader) Put(buffer []byte) {
   w := writer{buf: buffer}
   w.PutUint32(h.Size)
   w.PutBytes(h.Type[:])
}

// --- MDAT ---
type MdatBox struct {
   Header  *BoxHeader
   Payload []byte
}

func DecodeMdatBox(data []byte) (*MdatBox, error) {
   b := &MdatBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }
   b.Payload = data[8:b.Header.Size]
   return b, nil
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

type SidxBox struct {
   Header                   *BoxHeader
   Version                  byte
   Flags                    uint32
   ReferenceID              uint32
   Timescale                uint32
   EarliestPresentationTime uint64
   FirstOffset              uint64
   References               []SidxReference
}

func DecodeSidxBox(data []byte) (*SidxBox, error) {
   b := &SidxBox{}
   var err error
   b.Header, err = DecodeBoxHeader(data)
   if err != nil {
      return nil, err
   }

   if len(data) < 20 { // 8 byte header + 12 bytes of fields before version check
      return nil, errors.New("sidx box too short")
   }

   p := parser{data: data, offset: 8}
   versionAndFlags := p.Uint32()
   b.Version = byte(versionAndFlags >> 24)
   b.Flags = versionAndFlags & 0x00FFFFFF
   b.ReferenceID = p.Uint32()
   b.Timescale = p.Uint32()

   if b.Version == 0 {
      if len(data) < p.offset+8 {
         return nil, errors.New("sidx v0 box too short")
      }
      b.EarliestPresentationTime = uint64(p.Uint32())
      b.FirstOffset = uint64(p.Uint32())
   } else {
      if len(data) < p.offset+16 {
         return nil, errors.New("sidx v1 box too short")
      }
      b.EarliestPresentationTime = p.Uint64()
      b.FirstOffset = p.Uint64()
   }

   if len(data) < p.offset+4 {
      return nil, errors.New("sidx box too short for reference_count")
   }
   _ = p.Uint16() // reserved
   referenceCount := p.Uint16()

   if len(data)-p.offset < int(referenceCount)*12 {
      return nil, errors.New("sidx box too short for declared references")
   }

   b.References = make([]SidxReference, referenceCount)
   for i := 0; i < int(referenceCount); i++ {
      val1 := p.Uint32()
      b.References[i].ReferenceType = (val1 >> 31) == 1
      b.References[i].ReferencedSize = val1 & 0x7FFFFFFF

      b.References[i].SubsegmentDuration = p.Uint32()

      val2 := p.Uint32()
      b.References[i].StartsWithSAP = (val2 >> 31) == 1
      b.References[i].SAPType = uint8((val2 >> 28) & 0x07)
      b.References[i].SAPDeltaTime = val2 & 0x0FFFFFFF
   }
   return b, nil
}

func FindSidx(boxes []Box) (*SidxBox, bool) {
   for _, box := range boxes {
      if box.Sidx != nil {
         return box.Sidx, true
      }
   }
   return nil, false
}

// --- SIDX ---
type SidxReference struct {
   ReferenceType      bool
   ReferencedSize     uint32
   SubsegmentDuration uint32
   StartsWithSAP      bool
   SAPType            uint8
   SAPDeltaTime       uint32
}

// --- READING HELPER ---

type parser struct {
   data   []byte
   offset int
}

func (p *parser) Byte() byte {
   val := p.data[p.offset]
   p.offset++
   return val
}

func (p *parser) Bytes(n int) []byte {
   val := p.data[p.offset : p.offset+n]
   p.offset += n
   return val
}

func (p *parser) Int32() int32 {
   val := int32(binary.BigEndian.Uint32(p.data[p.offset:]))
   p.offset += 4
   return val
}

func (p *parser) Uint16() uint16 {
   val := binary.BigEndian.Uint16(p.data[p.offset:])
   p.offset += 2
   return val
}

func (p *parser) Uint32() uint32 {
   val := binary.BigEndian.Uint32(p.data[p.offset:])
   p.offset += 4
   return val
}

func (p *parser) Uint64() uint64 {
   val := binary.BigEndian.Uint64(p.data[p.offset:])
   p.offset += 8
   return val
}

// --- WRITING HELPER ---

type writer struct {
   buf    []byte
   offset int
}

func (w *writer) PutByte(data byte) {
   w.buf[w.offset] = data
   w.offset++
}

func (w *writer) PutBytes(data []byte) {
   copy(w.buf[w.offset:], data)
   w.offset += len(data)
}

func (w *writer) PutUint16(val uint16) {
   binary.BigEndian.PutUint16(w.buf[w.offset:], val)
   w.offset += 2
}

func (w *writer) PutUint32(val uint32) {
   binary.BigEndian.PutUint32(w.buf[w.offset:], val)
   w.offset += 4
}

func (w *writer) PutUint64(val uint64) {
   binary.BigEndian.PutUint64(w.buf[w.offset:], val)
   w.offset += 8
}

// core.go
