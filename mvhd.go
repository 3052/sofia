package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

type MvhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   // The rest of the fields (Rate, Volume, Matrix, NextTrackID, etc.)
   // are stored here to preserve them without defining specific structs.
   // This is usually 80 bytes (Rate(4)+Vol(2)+Rsrv(10)+Matrix(36)+PreDef(24)+NextID(4))
   RemainingData []byte
}

func (b *MvhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return fmt.Errorf("mvhd box too small")
   }

   b.Version = data[8]
   copy(b.Flags[:], data[9:12])

   offset := 12
   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mvhd v1 too short")
      }
      b.CreationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.ModificationTime = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = binary.BigEndian.Uint64(data[offset : offset+8])
      offset += 8
   } else { // Version 0
      if len(data) < 32 {
         return errors.New("mvhd v0 too short")
      }
      b.CreationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.ModificationTime = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
      b.Timescale = binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.Duration = uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
      offset += 4
   }

   // Copy remaining bytes (Rate, Volume, Matrix, NextTrackID)
   if offset < int(b.Header.Size) {
      b.RemainingData = make([]byte, int(b.Header.Size)-offset)
      copy(b.RemainingData, data[offset:b.Header.Size])
   }

   return nil
}

// SetDuration updates the duration.
// It automatically upgrades to Version 1 if the duration exceeds 32 bits.
func (b *MvhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MvhdBox) Encode() []byte {
   // Calculate total size based on version
   var baseSize uint32
   if b.Version == 1 {
      baseSize = 44
   } else {
      baseSize = 32
   }

   totalSize := baseSize + uint32(len(b.RemainingData))
   buf := make([]byte, totalSize)

   binary.BigEndian.PutUint32(buf[0:4], totalSize)
   copy(buf[4:8], b.Header.Type[:])
   buf[8] = b.Version
   copy(buf[9:12], b.Flags[:])

   offset := 12
   if b.Version == 1 {
      binary.BigEndian.PutUint64(buf[offset:offset+8], b.CreationTime)
      offset += 8
      binary.BigEndian.PutUint64(buf[offset:offset+8], b.ModificationTime)
      offset += 8
      binary.BigEndian.PutUint32(buf[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint64(buf[offset:offset+8], b.Duration)
      offset += 8
   } else {
      binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(b.CreationTime))
      offset += 4
      binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(b.ModificationTime))
      offset += 4
      binary.BigEndian.PutUint32(buf[offset:offset+4], b.Timescale)
      offset += 4
      binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(b.Duration))
      offset += 4
   }

   copy(buf[offset:], b.RemainingData)
   b.Header.Size = totalSize
   return buf
}
