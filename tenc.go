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

   // A 'tenc' box is a "Full Box".
   // The KID starts after the main header (8 bytes), the full box header (4 bytes),
   // a 24-bit reserved field (3 bytes), and an 8-bit IV size field (1 byte).
   // So, the KID offset is 8 + 4 + 3 + 1 = 16.
   const kidStart = 16
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
