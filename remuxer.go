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
   OnSample            func(sample []byte, encInfo *SampleEncryptionInfo)
}

// RemuxSample represents the minimal sample information needed for remuxing.
type RemuxSample struct {
   Size     uint32
   Duration uint32
   IsSync   bool
}

func (r *Remuxer) Initialize(initSegment []byte) error {
   if r.Moov != nil {
      return errors.New("already initialized")
   }
   if r.Writer == nil {
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
   r.Moov = moovPtr
   if _, ok := r.Moov.Trak(); !ok {
      return errors.New("no trak found")
   }
   r.mdatStartOffset, _ = r.Writer.Seek(0, io.SeekCurrent)
   mdatHeader := make([]byte, 16)
   binary.BigEndian.PutUint32(mdatHeader[0:4], 1)
   copy(mdatHeader[4:8], []byte("mdat"))
   if _, err := r.Writer.Write(mdatHeader); err != nil {
      return err
   }
   return nil
}

func (r *Remuxer) AddSegment(segmentData []byte) error {
   if r.Moov == nil {
      return errors.New("must call Initialize")
   }
   r.segmentCount++
   boxes, err := Parse(segmentData)
   if err != nil {
      return fmt.Errorf("parsing segment %d: %w", r.segmentCount, err)
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
            if err := r.processFragment(pendingMoof, box.Mdat); err != nil {
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
