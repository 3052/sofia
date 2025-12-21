package sofia

import "encoding/binary"

// --- STBL ---
type StblChild struct {
   Stsd *StsdBox
   Raw  []byte
}

type StblBox struct {
   Header   BoxHeader
   Children []StblChild
}

func (b *StblBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child StblChild
      switch string(h.Type[:]) {
      case "stsd":
         var stsd StsdBox
         if err := stsd.Parse(content); err != nil {
            return err
         }
         child.Stsd = &stsd
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *StblBox) Encode() []byte {
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Stsd != nil {
         buf = append(buf, child.Stsd.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
   return buf
}

func (b *StblBox) Stsd() (*StsdBox, bool) {
   for _, child := range b.Children {
      if child.Stsd != nil {
         return child.Stsd, true
      }
   }
   return nil, false
}

// --- STTS ---
type SttsEntry struct {
   SampleCount    uint32
   SampleDuration uint32
}

type SttsBox struct {
   Header  BoxHeader
   Entries []SttsEntry
}

func (b *SttsBox) Encode() []byte {
   buf := make([]byte, 16)
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Entries)))
   tmp := make([]byte, 8)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tmp[0:4], entry.SampleCount)
      binary.BigEndian.PutUint32(tmp[4:8], entry.SampleDuration)
      buf = append(buf, tmp...)
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 't', 's'}
   b.Header.Put(buf)
   return buf
}

func buildStts(samples []RemuxSample) []byte {
   if len(samples) == 0 {
      return nil
   }
   var entries []SttsEntry
   currDur := samples[0].Duration
   currCnt := uint32(0)
   for _, s := range samples {
      if s.Duration == currDur {
         currCnt++
      } else {
         entries = append(entries, SttsEntry{currCnt, currDur})
         currDur = s.Duration
         currCnt = 1
      }
   }
   entries = append(entries, SttsEntry{currCnt, currDur})
   box := SttsBox{Entries: entries}
   return box.Encode()
}

// --- STSZ ---
type StszBox struct {
   Header      BoxHeader
   SampleSize  uint32
   SampleCount uint32
   EntrySizes  []uint32
}

func (b *StszBox) Encode() []byte {
   buf := make([]byte, 20)
   binary.BigEndian.PutUint32(buf[12:16], b.SampleSize)
   binary.BigEndian.PutUint32(buf[16:20], b.SampleCount)
   tmp := make([]byte, 4)
   for _, size := range b.EntrySizes {
      binary.BigEndian.PutUint32(tmp, size)
      buf = append(buf, tmp...)
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 's', 'z'}
   b.Header.Put(buf)
   return buf
}

func buildStsz(samples []RemuxSample) []byte {
   entries := make([]uint32, len(samples))
   for i, s := range samples {
      entries[i] = s.Size
   }
   box := StszBox{SampleSize: 0, SampleCount: uint32(len(samples)), EntrySizes: entries}
   return box.Encode()
}

// --- STSC ---
type StscEntry struct {
   FirstChunk             uint32
   SamplesPerChunk        uint32
   SampleDescriptionIndex uint32
}

type StscBox struct {
   Header  BoxHeader
   Entries []StscEntry
}

func (b *StscBox) Encode() []byte {
   buf := make([]byte, 16)
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Entries)))
   tmp := make([]byte, 12)
   for _, entry := range b.Entries {
      binary.BigEndian.PutUint32(tmp[0:4], entry.FirstChunk)
      binary.BigEndian.PutUint32(tmp[4:8], entry.SamplesPerChunk)
      binary.BigEndian.PutUint32(tmp[8:12], entry.SampleDescriptionIndex)
      buf = append(buf, tmp...)
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 's', 'c'}
   b.Header.Put(buf)
   return buf
}

func buildStsc(counts []uint32) []byte {
   var entries []StscEntry
   chunkIdx := uint32(1)
   for _, cnt := range counts {
      if len(entries) > 0 {
         last := &entries[len(entries)-1]
         if last.SamplesPerChunk == cnt {
            chunkIdx++
            continue
         }
      }
      entries = append(entries, StscEntry{chunkIdx, cnt, 1})
      chunkIdx++
   }
   box := StscBox{Entries: entries}
   return box.Encode()
}

// --- STCO ---
type StcoBox struct {
   Header  BoxHeader
   Offsets []uint32
}

func (b *StcoBox) Encode() []byte {
   buf := make([]byte, 16)
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Offsets)))
   tmp := make([]byte, 4)
   for _, o := range b.Offsets {
      binary.BigEndian.PutUint32(tmp, o)
      buf = append(buf, tmp...)
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 'c', 'o'}
   b.Header.Put(buf)
   return buf
}

func buildStco(offsets []uint64) []byte {
   entries := make([]uint32, len(offsets))
   for i, o := range offsets {
      entries[i] = uint32(o)
   }
   box := StcoBox{Offsets: entries}
   return box.Encode()
}

// --- STSS ---
type StssBox struct {
   Header  BoxHeader
   Indices []uint32
}

func (b *StssBox) Encode() []byte {
   buf := make([]byte, 16)
   binary.BigEndian.PutUint32(buf[12:16], uint32(len(b.Indices)))
   tmp := make([]byte, 4)
   for _, idx := range b.Indices {
      binary.BigEndian.PutUint32(tmp, idx)
      buf = append(buf, tmp...)
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Type = [4]byte{'s', 't', 's', 's'}
   b.Header.Put(buf)
   return buf
}

func buildStss(samples []RemuxSample) []byte {
   var indices []uint32
   for i, s := range samples {
      if s.IsSync {
         indices = append(indices, uint32(i+1))
      }
   }
   // If all samples are sync samples, stss is not required (and omitting it saves space).
   // However, if there is at least one non-sync sample, we MUST provide the table.
   if len(indices) == len(samples) {
      return nil
   }
   box := StssBox{Indices: indices}
   return box.Encode()
}
