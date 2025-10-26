package mp4

import "fmt"

// TencBox represents the 'tenc' box (Track Encryption Box).
type TencBox struct {
   Header     BoxHeader
   RawData    []byte // Stores the original box data for a perfect round trip
   DefaultKID [16]byte
}

// ParseTenc parses the 'tenc' box from a byte slice.
func ParseTenc(data []byte) (TencBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TencBox{}, err
   }
   var tenc TencBox
   tenc.Header = header
   tenc.RawData = data[:header.Size] // Store the original data

   // Also parse the fields needed for decryption
   if len(data) < 31 {
      return TencBox{}, fmt.Errorf("tenc box is too small to contain KID: %d bytes", len(data))
   }
   copy(tenc.DefaultKID[:], data[15:31])

   return tenc, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TencBox) Encode() []byte {
   return b.RawData
}
