package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

type MdhdBox struct {
   Header           BoxHeader
   Version          byte
   Flags            [3]byte
   CreationTime     uint64
   ModificationTime uint64
   Timescale        uint32
   Duration         uint64
   Language         [2]byte
   Quality          [2]byte
}

func (b *MdhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }

   if len(data) < 12 {
      return fmt.Errorf("mdhd box too small")
   }

   b.Version = data[8]
   copy(b.Flags[:], data[9:12])

   offset := 12
   if b.Version == 1 {
      if len(data) < 44 {
         return errors.New("mdhd v1 too short")
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
         return errors.New("mdhd v0 too short")
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

   if len(data) < offset+4 {
      return errors.New("mdhd truncated at language/quality")
   }
   copy(b.Language[:], data[offset:offset+2])
   copy(b.Quality[:], data[offset+2:offset+4])

   return nil
}

// SetDuration updates the duration.
// It automatically upgrades to Version 1 if the duration exceeds 32 bits.
func (b *MdhdBox) SetDuration(duration uint64) {
   b.Duration = duration
   if b.Duration > 0xFFFFFFFF {
      b.Version = 1
   }
}

func (b *MdhdBox) Encode() []byte {
   var size uint32
   if b.Version == 1 {
      size = 44
   } else {
      size = 32
   }

   buf := make([]byte, size)

   binary.BigEndian.PutUint32(buf[0:4], size)
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

   copy(buf[offset:offset+2], b.Language[:])
   copy(buf[offset+2:offset+4], b.Quality[:])

   b.Header.Size = size
   return buf
}
