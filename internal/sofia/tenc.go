package mp4

import "fmt"

// TencBox represents the 'tenc' box (Track Encryption Box).
type TencBox struct {
   Header     BoxHeader
   RawData    []byte // Stores the original box data for a perfect round trip
   DefaultKID [16]byte
}

// Parse parses the 'tenc' box from a byte slice.
func (b *TencBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size] // Store the original data

   // Also parse the fields needed for decryption.
   // A tenc box is a "full box" (version+flags), taking 12 bytes total for the header part.
   // The KID starts after 1 byte (isProtected) and 1 byte (IV size).
   // So, the KID offset is 8 (box header) + 4 (full box header) + 1 + 1 = 14.
   kidStart := 14
   if len(data) < kidStart+16 {
      return fmt.Errorf("tenc box is too small to contain KID: %d bytes", len(data))
   }
   // Correctly copy from offset 14 to 30.
   copy(b.DefaultKID[:], data[kidStart:kidStart+16])

   return nil
}

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *TencBox) Encode() []byte {
   return b.RawData
}
