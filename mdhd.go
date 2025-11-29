package sofia

import (
   "encoding/binary"
   "errors"
   "fmt"
)

type MdhdBox struct {
   Header    BoxHeader
   RawData   []byte
   Version   byte
   Timescale uint32
   Duration  uint64
}

func (b *MdhdBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]

   if len(data) < 12 {
      return fmt.Errorf("mdhd box is too small: %d bytes", len(data))
   }

   b.Version = data[8]
   if b.Version == 1 {
      if len(data) < 36 {
         return fmt.Errorf("mdhd version 1 box is too small: %d bytes", len(data))
      }
      b.Timescale = binary.BigEndian.Uint32(data[28:32])
      b.Duration = binary.BigEndian.Uint64(data[32:40])
   } else {
      if len(data) < 24 {
         return fmt.Errorf("mdhd version 0 box is too small: %d bytes", len(data))
      }
      b.Timescale = binary.BigEndian.Uint32(data[20:24])
      b.Duration = uint64(binary.BigEndian.Uint32(data[24:28]))
   }

   return nil
}

// SetDuration updates the duration field directly in the RawData.
func (b *MdhdBox) SetDuration(duration uint64) error {
   if b.Version == 1 {
      // Version 1: Duration is at offset 32 (8 bytes)
      if len(b.RawData) < 40 {
         return errors.New("mdhd raw data too short for v1 duration")
      }
      binary.BigEndian.PutUint64(b.RawData[32:], duration)
   } else {
      // Version 0: Duration is at offset 24 (4 bytes)
      if len(b.RawData) < 28 {
         return errors.New("mdhd raw data too short for v0 duration")
      }
      if duration > 0xFFFFFFFF {
         return errors.New("duration overflows 32-bit mdhd field")
      }
      binary.BigEndian.PutUint32(b.RawData[24:], uint32(duration))
   }
   b.Duration = duration // Update struct field for consistency
   return nil
}
