package sofia

import "fmt"

// TencBox represents the 'tenc' box (Track Encryption Box).
type TencBox struct {
   Header     BoxHeader
   RawData    []byte // Stores the original box data for a perfect round trip
   DefaultKID [16]byte
}

// Parse parses the 'tenc' box from a byte slice.
func (b *TencBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size] // Store the original data

   kidStart := 14
   if len(data) < kidStart+16 {
      return fmt.Errorf("tenc box is too small to contain KID: %d bytes", len(data))
   }
   copy(b.DefaultKID[:], data[kidStart:kidStart+16])

   return nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TencBox) Encode() []byte {
   return b.RawData
}
