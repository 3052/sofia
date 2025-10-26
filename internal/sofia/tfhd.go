package mp4

import "encoding/binary"

// TfhdBox represents the 'tfhd' box (Track Fragment Header Box).
type TfhdBox struct {
   Header  BoxHeader
   RawData []byte // Stores the original box data for a perfect round trip
   Flags   uint32
   TrackID uint32
}

// ParseTfhd parses the 'tfhd' box from a byte slice.
func ParseTfhd(data []byte) (TfhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TfhdBox{}, err
   }
   var tfhd TfhdBox
   tfhd.Header = header
   tfhd.RawData = data[:header.Size] // Store the original data

   // Also parse the fields needed for decryption
   tfhd.Flags = binary.BigEndian.Uint32(data[8:12]) & 0x00FFFFFF
   tfhd.TrackID = binary.BigEndian.Uint32(data[12:16])
   return tfhd, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TfhdBox) Encode() []byte {
   return b.RawData
}
