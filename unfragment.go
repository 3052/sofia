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

   moof := FindMoofPtr(boxes)
   mdat := FindMdatPtr(boxes)

   if moof == nil || mdat == nil {
      return nil
   }

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

   mdatOffset := 0

   for _, child := range traf.Children {
      if child.Trun != nil {
         trun := child.Trun
         for _, s := range trun.Samples {
            si := UnfragSample{Duration: defDur, Size: defSize}
            if (trun.Flags & 0x000100) != 0 {
               si.Duration = s.Duration
            }
            if (trun.Flags & 0x000200) != 0 {
               si.Size = s.Size
            }

            // Store original size to ensure parser stays in sync with mdat payload
            // even if the user modifies si.Size in the callback.
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

   // Capture the end of mdat payload before writing moov
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

   trak, _ := u.Moov.Trak()
   mdia, _ := trak.Mdia()
   minf, _ := mdia.Minf()
   stbl, _ := minf.Stbl()

   mdhd, ok := mdia.Mdhd()
   if !ok {
      return errors.New("missing mdhd")
   }
   // Update mdhd directly using the new method
   if err := mdhd.SetDuration(totalDuration); err != nil {
      return err
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
