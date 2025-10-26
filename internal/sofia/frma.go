package mp4

import "fmt"

// FrmaBox represents the 'frma' box (Original Format Box).
type FrmaBox struct {
   Header     BoxHeader
   RawData    []byte // Stores the original box data for a perfect round trip
   DataFormat [4]byte
}

// ParseFrma parses the 'frma' box from a byte slice.
func ParseFrma(data []byte) (FrmaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return FrmaBox{}, err
   }
   var frma FrmaBox
   frma.Header = header
   frma.RawData = data[:header.Size]

   // The dataFormat is the 4 bytes immediately following the box header.
   if len(data) < 12 {
      return FrmaBox{}, fmt.Errorf("frma box is too small: %d bytes", len(data))
   }
   copy(frma.DataFormat[:], data[8:12])

   return frma, nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *FrmaBox) Encode() []byte {
   return b.RawData
}
