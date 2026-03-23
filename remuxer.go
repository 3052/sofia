// remuxer.go
package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

type Remuxer struct {
   Writer              io.WriteSeeker
   Moov                *MoovBox
   samples             []RemuxSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   segmentCount        int
   OnSample            func(data []byte, sample *SencSample)
}

type RemuxSample struct {
   Size                  uint32
   Duration              uint32
   IsSync                bool
   CompositionTimeOffset int32
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
   r.mdatStartOffset, _ = r.Writer.Seek(0, io.SeekCurrent)
   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))
   _, err = r.Writer.Write(mdatHeader)
   return err
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

func (r *Remuxer) processFragment(moof *MoofBox, mdat *MdatBox) error {
   traf := moof.Traf
   if traf == nil {
      return nil
   }
   tfhd := traf.Tfhd
   if tfhd == nil {
      return nil
   }
   senc := traf.Senc
   sencIndex := 0
   var newSamples []RemuxSample
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize
   defFlags := tfhd.DefaultSampleFlags
   mdatOffset := 0
   for _, trun := range traf.Trun {
      for i, sample := range trun.Samples {
         remuxSample := RemuxSample{
            Duration:              defDur,
            Size:                  defSize,
            IsSync:                true,
            CompositionTimeOffset: 0,
         }
         currentFlags := defFlags
         if i == 0 && (trun.Flags&0x000004) != 0 {
            currentFlags = trun.FirstSampleFlags
         }
         if (trun.Flags & 0x000400) != 0 {
            currentFlags = sample.Flags
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
         originalSize := int(remuxSample.Size)
         if mdatOffset+originalSize > len(mdat.Payload) {
            return errors.New("mdat payload too short for samples")
         }
         sampleData := mdat.Payload[mdatOffset : mdatOffset+originalSize]
         var encInfo *SencSample
         if senc != nil && sencIndex < len(senc.Samples) {
            encInfo = &senc.Samples[sencIndex]
            sencIndex++
         }
         if r.OnSample != nil {
            r.OnSample(sampleData, encInfo)
         }
         newSamples = append(newSamples, remuxSample)
         mdatOffset += originalSize
      }
   }

   if len(newSamples) == 0 {
      return nil
   }
   currentPos, _ := r.Writer.Seek(0, io.SeekCurrent)
   r.chunkOffsets = append(r.chunkOffsets, uint64(currentPos))
   if _, err := r.Writer.Write(mdat.Payload); err != nil {
      return err
   }
   r.samples = append(r.samples, newSamples...)
   r.segmentSampleCounts = append(r.segmentSampleCounts, uint32(len(newSamples)))
   return nil
}

func (r *Remuxer) Finish() error {
   if r.Moov == nil {
      return errors.New("not initialized")
   }
   mdatEndOffset, _ := r.Writer.Seek(0, io.SeekCurrent)
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
      return err
   }
   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := r.Writer.Write(sizeBuf[:]); err != nil {
      return err
   }
   r.Writer.Seek(0, io.SeekEnd)
   return nil
}
