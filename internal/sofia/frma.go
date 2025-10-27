package mp4

import "fmt"

// FrmaBox represents the 'frma' box (Original Format Box).
type FrmaBox struct {
   Header     BoxHeader
   RawData    []byte // Stores the original box data for a perfect round trip
   DataFormat [4]byte
}

// Parse parses the 'frma' box from a byte slice.
func (b *FrmaBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]

   // The dataFormat is the 4 bytes immediately following the box header.
   if len(data) < 12 {
      return fmt.Errorf("frma box is too small: %d bytes", len(data))
   }
   copy(b.DataFormat[:], data[8:12])

   return nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *FrmaBox) Encode() []byte {
   return b.RawData
}
