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

// ParseMdhd parses the 'mdhd' box from a byte slice.
func ParseMdhd(data []byte) (MdhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdhdBox{}, err
   }
   var mdhd MdhdBox
   mdhd.Header = header
   mdhd.RawData = data[:header.Size]

   if len(data) < 12 {
      return MdhdBox{}, fmt.Errorf("mdhd box is too small: %d bytes", len(data))
   }

   mdhd.Version = data[8]
   if mdhd.Version == 1 {
      // 64-bit duration version
      if len(data) < 36 {
         return MdhdBox{}, fmt.Errorf("mdhd version 1 box is too small: %d bytes", len(data))
      }
      mdhd.Timescale = binary.BigEndian.Uint32(data[28:32])
      mdhd.Duration = binary.BigEndian.Uint64(data[32:40])
   } else {
      // 32-bit duration version
      if len(data) < 24 {
         return MdhdBox{}, fmt.Errorf("mdhd version 0 box is too small: %d bytes", len(data))
      }
      mdhd.Timescale = binary.BigEndian.Uint32(data[20:24])
      mdhd.Duration = uint64(binary.BigEndian.Uint32(data[24:28]))
   }

   return mdhd, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *MdhdBox) Encode() []byte {
   return b.RawData
}
