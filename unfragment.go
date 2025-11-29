package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
   "io"
   "log"
)

type Unfragmenter struct {
   dst                 io.WriteSeeker
   moov                *MoovBox
   Samples             []UnfragSample
   chunkOffsets        []uint64
   segmentSampleCounts []uint32
   mdatStartOffset     int64
   payloadWritten      uint64
   initialized         bool
   segmentCount        int
}

type UnfragSample struct {
   Size     uint32
   Duration uint32
}

func NewUnfragmenter(dst io.WriteSeeker) *Unfragmenter {
   return &Unfragmenter{dst: dst}
}

func (u *Unfragmenter) Initialize(initSegment []byte) error {
   if u.initialized {
      return errors.New("already initialized")
   }

   log.Println("[Unfrag] Initializing...")
   boxes, err := Parse(initSegment)
   if err != nil {
      return fmt.Errorf("parsing init: %w", err)
   }

   moovPtr, ok := FindMoov(boxes)
   if !ok {
      return errors.New("no moov found")
   }
   u.moov = moovPtr

   if _, ok := u.moov.Trak(); !ok {
      return errors.New("no trak found")
   }

   u.mdatStartOffset, _ = u.dst.Seek(0, io.SeekCurrent)

   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))
   if _, err := u.dst.Write(mdatHeader); err != nil {
      return err
   }

   u.initialized = true
   return nil
}

func (u *Unfragmenter) AddSegment(segmentData []byte) error {
   if !u.initialized {
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

   var newSamples []UnfragSample
   defDur := tfhd.DefaultSampleDuration
   defSize := tfhd.DefaultSampleSize

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
            newSamples = append(newSamples, si)
         }
      }
   }

   if len(newSamples) == 0 {
      return nil
   }

   currentPos, _ := u.dst.Seek(0, io.SeekCurrent)
   u.chunkOffsets = append(u.chunkOffsets, uint64(currentPos))

   n, err := u.dst.Write(mdat.Payload)
   if err != nil {
      return err
   }
   u.payloadWritten += uint64(n)

   u.Samples = append(u.Samples, newSamples...)
   u.segmentSampleCounts = append(u.segmentSampleCounts, uint32(len(newSamples)))

   return nil
}

func (u *Unfragmenter) Finish() error {
   if !u.initialized {
      return errors.New("not initialized")
   }

   log.Println("[Unfrag] Finishing...")

   var totalDuration uint64
   for _, s := range u.Samples {
      totalDuration += uint64(s.Duration)
   }

   stts := buildStts(u.Samples)
   stsz := buildStsz(u.Samples)
   stsc := buildStsc(u.segmentSampleCounts)
   offsetBox := buildStco(u.chunkOffsets)

   trak, _ := u.moov.Trak()
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

   u.moov.RemoveMvex()

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

   moovBytes := u.moov.Encode()
   if _, err := u.dst.Write(moovBytes); err != nil {
      return err
   }

   if _, err := u.dst.Seek(u.mdatStartOffset+8, io.SeekStart); err != nil {
      return err
   }

   finalMdatSize := uint64(16) + u.payloadWritten
   var sizeBuf [8]byte
   binary.BigEndian.PutUint64(sizeBuf[:], finalMdatSize)
   if _, err := u.dst.Write(sizeBuf[:]); err != nil {
      return err
   }

   u.dst.Seek(0, io.SeekEnd)
   return nil
}
