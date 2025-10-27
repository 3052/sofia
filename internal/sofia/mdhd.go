package mp4

import (
   "encoding/binary"
   "fmt"
)

// MdhdBox represents the 'mdhd' box (Media Header Box).
type MdhdBox struct {
   Header    BoxHeader
   RawData   []byte
   Version   byte
   Timescale uint32
   Duration  uint64
}

// Parse parses the 'mdhd' box from a byte slice.
func (b *MdhdBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]

   if len(data) < 12 {
      return fmt.Errorf("mdhd box is too small: %d bytes", len(data))
   }

   b.Version = data[8]
   if b.Version == 1 {
      // 64-bit duration version
      if len(data) < 36 {
         return fmt.Errorf("mdhd version 1 box is too small: %d bytes", len(data))
      }
      b.Timescale = binary.BigEndian.Uint32(data[28:32])
      b.Duration = binary.BigEndian.Uint64(data[32:40])
   } else {
      // 32-bit duration version
      if len(data) < 24 {
         return fmt.Errorf("mdhd version 0 box is too small: %d bytes", len(data))
      }
      b.Timescale = binary.BigEndian.Uint32(data[20:24])
      b.Duration = uint64(binary.BigEndian.Uint32(data[24:28]))
   }

   return nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *MdhdBox) Encode() []byte {
   return b.RawData
}
