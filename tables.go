package sofia

import "errors"

// encodeTable builds a table box from raw 4-byte entries. count is the
// entry count written into the box, which differs from len(entries) when
// one entry spans several words (stts, ctts, stsc).
func encodeTable(name string, count int, entries []uint32) []byte {
   buffer := make([]byte, 16+len(entries)*4)
   w := writer{buf: buffer}
   w.offset = 8
   w.PutUint32(0) // version and flags
   w.PutUint32(uint32(count))
   for _, entry := range entries {
      w.PutUint32(entry)
   }
   header := &BoxHeader{Size: uint32(len(buffer))}
   copy(header.Type[:], name)
   header.Put(buffer)
   return buffer
}

// encodeTable64 is encodeTable with 8-byte entries.
func encodeTable64(name string, entries []uint64) []byte {
   buffer := make([]byte, 16+len(entries)*8)
   w := writer{buf: buffer}
   w.offset = 8
   w.PutUint32(0) // version and flags
   w.PutUint32(uint32(len(entries)))
   for _, entry := range entries {
      w.PutUint64(entry)
   }
   header := &BoxHeader{Size: uint32(len(buffer))}
   copy(header.Type[:], name)
   header.Put(buffer)
   return buffer
}

// The sample tables of a progressive MP4. stbl is the container; its child
// boxes map to RemuxSample fields (stts durations, ctts composition time
// offsets, stsz sizes, stss sync flags) and to chunk locations (stsc
// counts, stco/co64 offsets). The table boxes store only their entries:
// headers exist only in encoded bytes and are rebuilt by the encoders.

// tableU32 decodes the common table shape — version/flags, entry count,
// then count 4-byte entries — and returns the raw entries.
func tableU32(data []byte) ([]uint32, error) {
   if len(data) < 16 {
      return nil, errors.New("box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32() // version and flags
   count := p.Uint32()
   if int(count) > (len(data)-p.offset)/4 {
      return nil, errors.New("box too short for declared entries")
   }
   entries := make([]uint32, count)
   for i := range entries {
      entries[i] = p.Uint32()
   }
   return entries, nil
}

// tableU64 is tableU32 with 8-byte entries.
func tableU64(data []byte) ([]uint64, error) {
   if len(data) < 16 {
      return nil, errors.New("box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32() // version and flags
   count := p.Uint32()
   if int(count) > (len(data)-p.offset)/8 {
      return nil, errors.New("box too short for declared entries")
   }
   entries := make([]uint64, count)
   for i := range entries {
      entries[i] = p.Uint64()
   }
   return entries, nil
}

// --- CO64 ---
type Co64Box struct {
   Offsets []uint64
}

func DecodeCo64Box(data []byte) (*Co64Box, error) {
   offsets, err := tableU64(data)
   if err != nil {
      return nil, err
   }
   return &Co64Box{Offsets: offsets}, nil
}

func (b Co64Box) Encode() []byte {
   return encodeTable64("co64", b.Offsets)
}

// --- CTTS ---
type CttsBox struct {
   Entries []CttsEntry
}

func DecodeCttsBox(data []byte) (*CttsBox, error) {
   if len(data) < 16 {
      return nil, errors.New("box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32()      // version and flags
   count := p.Uint32() // entry count: each entry is 8 bytes
   if int(count) > (len(data)-p.offset)/8 {
      return nil, errors.New("ctts box too short for declared entries")
   }
   b := &CttsBox{Entries: make([]CttsEntry, count)}
   for i := range b.Entries {
      b.Entries[i].SampleCount = p.Uint32()
      b.Entries[i].SampleOffset = int32(p.Uint32())
   }
   return b, nil
}

func (b CttsBox) Encode() []byte {
   raw := make([]uint32, 2*len(b.Entries))
   for i, entry := range b.Entries {
      raw[2*i], raw[2*i+1] = entry.SampleCount, uint32(entry.SampleOffset)
   }
   return encodeTable("ctts", len(b.Entries), raw)
}

type CttsEntry struct {
   SampleCount  uint32
   SampleOffset int32
}

// --- STBL ---
type StblBox struct {
   Header      *BoxHeader
   Stsd        *StsdBox
   Stts        *SttsBox
   Ctts        *CttsBox
   Stsz        *StszBox
   Stsc        *StscBox
   Stco        *StcoBox
   Co64        *Co64Box
   Stss        *StssBox
   RawChildren [][]byte
}

func DecodeStblBox(data []byte) (*StblBox, error) {
   b := &StblBox{}
   var err error
   if b.Header, err = DecodeBoxHeader(data); err != nil {
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
      case "stsd":
         b.Stsd, err = DecodeStsdBox(content)
      case "stts":
         b.Stts, err = DecodeSttsBox(content)
      case "ctts":
         b.Ctts, err = DecodeCttsBox(content)
      case "stsz":
         b.Stsz, err = DecodeStszBox(content)
      case "stsc":
         b.Stsc, err = DecodeStscBox(content)
      case "stco":
         b.Stco, err = DecodeStcoBox(content)
      case "co64":
         b.Co64, err = DecodeCo64Box(content)
      case "stss":
         b.Stss, err = DecodeStssBox(content)
      default:
         b.RawChildren = append(b.RawChildren, content)
      }
      if err != nil {
         return nil, err
      }
      offset += boxSize
   }
   return b, nil
}

func (b *StblBox) Encode() []byte {
   buffer := make([]byte, 8)
   if b.Stsd != nil {
      buffer = append(buffer, b.Stsd.Encode()...)
   }
   for _, child := range b.RawChildren {
      buffer = append(buffer, child...)
   }
   b.Header.Size = uint32(len(buffer))
   b.Header.Put(buffer)
   return buffer
}

// --- STCO ---
type StcoBox struct {
   Offsets []uint32
}

func DecodeStcoBox(data []byte) (*StcoBox, error) {
   offsets, err := tableU32(data)
   if err != nil {
      return nil, err
   }
   return &StcoBox{Offsets: offsets}, nil
}

func (b StcoBox) Encode() []byte {
   return encodeTable("stco", len(b.Offsets), b.Offsets)
}

// --- STSC ---
type StscBox struct {
   Entries []StscEntry
}

func DecodeStscBox(data []byte) (*StscBox, error) {
   if len(data) < 16 {
      return nil, errors.New("box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32() // version and flags
   count := p.Uint32()
   if int(count) > (len(data)-p.offset)/12 {
      return nil, errors.New("box too short for declared entries")
   }
   b := &StscBox{Entries: make([]StscEntry, count)}
   for i := range b.Entries {
      b.Entries[i].FirstChunk = p.Uint32()
      b.Entries[i].SamplesPerChunk = p.Uint32()
      b.Entries[i].SampleDescriptionIndex = p.Uint32()
   }
   return b, nil
}

func (b StscBox) Encode() []byte {
   raw := make([]uint32, 3*len(b.Entries))
   for i, entry := range b.Entries {
      raw[3*i], raw[3*i+1], raw[3*i+2] = entry.FirstChunk, entry.SamplesPerChunk, entry.SampleDescriptionIndex
   }
   return encodeTable("stsc", len(b.Entries), raw)
}

type StscEntry struct {
   FirstChunk             uint32
   SamplesPerChunk        uint32
   SampleDescriptionIndex uint32
}

// --- STSS ---
type StssBox struct {
   Indices []uint32
}

func DecodeStssBox(data []byte) (*StssBox, error) {
   indices, err := tableU32(data)
   if err != nil {
      return nil, err
   }
   return &StssBox{Indices: indices}, nil
}

func (b StssBox) Encode() []byte {
   return encodeTable("stss", len(b.Indices), b.Indices)
}

// --- STSZ ---
type StszBox struct {
   SampleSize  uint32
   SampleCount uint32
   EntrySizes  []uint32
}

func DecodeStszBox(data []byte) (*StszBox, error) {
   if len(data) < 20 {
      return nil, errors.New("stsz box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32() // version and flags
   b := &StszBox{}
   b.SampleSize = p.Uint32()
   b.SampleCount = p.Uint32()
   if b.SampleSize != 0 {
      return b, nil // constant sample size: no per-sample entries
   }
   if int(b.SampleCount) > (len(data)-p.offset)/4 {
      return nil, errors.New("stsz box too short for declared entries")
   }
   b.EntrySizes = make([]uint32, b.SampleCount)
   for i := range b.EntrySizes {
      b.EntrySizes[i] = p.Uint32()
   }
   return b, nil
}

func (b StszBox) Encode() []byte {
   buffer := make([]byte, 20+len(b.EntrySizes)*4)
   w := writer{buf: buffer}
   w.offset = 8
   w.PutUint32(0) // version and flags
   w.PutUint32(b.SampleSize)
   w.PutUint32(b.SampleCount)
   for _, size := range b.EntrySizes {
      w.PutUint32(size)
   }
   header := &BoxHeader{Size: uint32(len(buffer))}
   copy(header.Type[:], "stsz")
   header.Put(buffer)
   return buffer
}

// --- STTS ---
type SttsBox struct {
   Entries []SttsEntry
}

func DecodeSttsBox(data []byte) (*SttsBox, error) {
   if len(data) < 16 {
      return nil, errors.New("box too short")
   }
   p := parser{data: data, offset: 8}
   _ = p.Uint32()      // version and flags
   count := p.Uint32() // entry count: each entry is 8 bytes
   if int(count) > (len(data)-p.offset)/8 {
      return nil, errors.New("stts box too short for declared entries")
   }
   b := &SttsBox{Entries: make([]SttsEntry, count)}
   for i := range b.Entries {
      b.Entries[i].SampleCount = p.Uint32()
      b.Entries[i].SampleDuration = p.Uint32()
   }
   return b, nil
}

func (b SttsBox) Encode() []byte {
   raw := make([]uint32, 2*len(b.Entries))
   for i, entry := range b.Entries {
      raw[2*i], raw[2*i+1] = entry.SampleCount, entry.SampleDuration
   }
   return encodeTable("stts", len(b.Entries), raw)
}

type SttsEntry struct {
   SampleCount    uint32
   SampleDuration uint32
}

// tables.go
