package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
)

type Unfragmenter struct {
   Writer              io.WriteSeeker
   Moov                *MoovBox
   samples             []UnfragSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   segmentCount        int
   OnSample            func(sample []byte, encInfo *SampleEncryptionInfo)
   OnSampleInfo        func(*UnfragSample)
}

// UnfragSample represents the minimal sample information needed for unfragmenting.
type UnfragSample struct {
   Size     uint32
   Duration uint32
   IsSync   bool
}

func (u *Unfragmenter) Initialize(initSegment []byte) error {
   if u.Moov != nil {
      return errors.New("already initialized")
   }
   if u.Writer == nil {
      return errors.New("writer is nil")
   }

   boxes, err := Parse(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init: %w", err)
   }

   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found")
   }
   u.Moov = moovPtr

   if _, ok := u.Moov.Trak(); !ok {
      return errors.New("no trak found")
   }

   u.mdatStartOffset, _ = u.Writer.Seek(0, io.SeekCurrent)
   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))

   if _, err := u.Writer.Write(mdatHeader); err != nil {
      return err
   }

   return nil
}

func (u *Unfragmenter) AddSegment(segmentData []byte) error {
   if u.Moov == nil {
      return errors.New("must call Initialize")
   }
   u.segmentCount++

   boxes, err := Parse(segmentData)
   if err != nil {
      return fmt.Errorf("parsing segment %d: %w", u.segmentCount, err)
   }

   var pendingMoof *MoofBox
   foundPair := false

   for i, box := range boxes {
      if box.Moof != nil {
         pendingMoof = box.Moof
         continue
      }
      if box.Mdat != nil {
         if pendingMoof != nil {
            if err := u.processFragment(pendingMoof, box.Mdat); err != nil {
               return fmt.Errorf("processing fragment at box index %d: %w", i, err)
            }
            pendingMoof = nil
            foundPair = true
         }
      }
   }
   if !foundPair {
      return nil
   }

   return nil
}

func (u *Unfragmenter) processFragment(moof *MoofBox, mdat *MdatBox) error {
   traf, ok := moof.Traf()
   if !ok {
      return nil
   }

   tfhd := traf.Tfhd()
   if tfhd == nil {
      return nil
   }

   // Senc is optional and provides encryption info
   senc, _ := traf.Senc()
   sencIndex := 0

   var newSamples []UnfragSample
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize
   defFlags := tfhd.DefaultSampleFlags

   mdatOffset := 0

   for _, child := range traf.Children {
      if child.Trun != nil {
         trun := child.Trun
         for _, s := range trun.Samples {
            si := UnfragSample{Duration: defDur, Size: defSize, IsSync: true}
            currentFlags := defFlags

            if (trun.Flags & 0x000100) != 0 {
               si.Duration = s.Duration
            }
            if (trun.Flags & 0x000200) != 0 {
               si.Size = s.Size
            }
            if (trun.Flags & 0x000400) != 0 {
               currentFlags = s.Flags
            }

            // The flag 'sample_is_difference_sample' is typically 0x00010000.
            // If this bit is 1, the sample is NOT a sync sample (it's a difference sample).
            // If this bit is 0, the sample IS a sync sample.
            if (currentFlags & 0x00010000) != 0 {
               si.IsSync = false
            } else {
               si.IsSync = true
            }

            originalSize := int(si.Size)
            if mdatOffset+originalSize > len(mdat.Payload) {
               return errors.New("mdat payload too short for samples")
            }
            sampleData := mdat.Payload[mdatOffset : mdatOffset+originalSize]

            var encInfo *SampleEncryptionInfo
            if senc != nil && sencIndex < len(senc.Samples) {
               encInfo = &senc.Samples[sencIndex]
               sencIndex++
            }

            if u.OnSample != nil {
               u.OnSample(sampleData, encInfo)
            }
            if u.OnSampleInfo != nil {
               u.OnSampleInfo(&si)
            }

            newSamples = append(newSamples, si)
            mdatOffset += originalSize
         }
      }
   }

   if len(newSamples) == 0 {
      return nil
   }

   currentPos, _ := u.Writer.Seek(0, io.SeekCurrent)
   u.chunkOffsets = append(u.chunkOffsets, uint64(currentPos))

   if _, err := u.Writer.Write(mdat.Payload); err != nil {
      return err
   }

   u.samples = append(u.samples, newSamples...)
   u.segmentSampleCounts = append(u.segmentSampleCounts, uint32(len(newSamples)))
   return nil
}

func (u *Unfragmenter) Finish() error {
   if u.Moov == nil {
      return errors.New("not initialized")
   }

   mdatEndOffset, _ := u.Writer.Seek(0, io.SeekCurrent)
   finalMdatSize := uint64(mdatEndOffset - u.mdatStartOffset)

   var totalDuration uint64
   for _, s := range u.samples {
      totalDuration += uint64(s.Duration)
   }

   stts := buildStts(u.samples)
   stsz := buildStsz(u.samples)
   stsc := buildStsc(u.segmentSampleCounts)
   offsetBox := buildStco(u.chunkOffsets)
   stss := buildStss(u.samples) // Build the sync sample table

   trak, _ := u.Moov.Trak()
   mdia, _ := trak.Mdia()
   minf, _ := mdia.Minf()
   stbl, _ := minf.Stbl()
   mdhd, ok := mdia.Mdhd()
   if !ok {
      return errors.New("missing mdhd")
   }

   mdhd.SetDuration(totalDuration)

   // Update Mvhd duration if present
   if mvhd, ok := u.Moov.Mvhd(); ok {
      var movieDuration uint64
      // Convert duration from Track Time Scale to Movie Time Scale
      if mdhd.Timescale > 0 {
         movieDuration = (totalDuration * uint64(mvhd.Timescale)) / uint64(mdhd.Timescale)
      }
      mvhd.SetDuration(movieDuration)
   }

   u.Moov.RemoveMvex()

   var newChildren []StblChild
   if stsd, ok := stbl.Stsd(); ok {
      stsd.UnprotectAll()
      newChildren = append(newChildren, StblChild{Stsd: stsd})
   } else {
      return errors.New("missing stsd")
   }

   newChildren = append(newChildren, StblChild{Raw: stts})
   newChildren = append(newChildren, StblChild{Raw: stsz})
   newChildren = append(newChildren, StblChild{Raw: stsc})
   newChildren = append(newChildren, StblChild{Raw: offsetBox})

   // Only append stss if it was generated (i.e., we have mixed sync/non-sync frames)
   if stss != nil {
      newChildren = append(newChildren, StblChild{Raw: stss})
   }
   stbl.Children = newChildren

   moovBytes := u.Moov.Encode()
   if _, err := u.Writer.Write(moovBytes); err != nil {
      return err
   }

   if _, err := u.Writer.Seek(u.mdatStartOffset+8, io.SeekStart); err != nil {
      return err
   }

   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := u.Writer.Write(sizeBuf[:]); err != nil {
      return err
   }

   u.Writer.Seek(0, io.SeekEnd)
   return nil
}
