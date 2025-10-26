package mp4

import "fmt"

// TencBox represents the 'tenc' box (Track Encryption Box).
type TencBox struct {
   Header  BoxHeader
   Version byte
   Flags   [3]byte
   // Reserved fields omitted for simplicity
   DefaultPerSampleIVSize byte
   DefaultKID             [16]byte
}

// ParseTenc parses the 'tenc' box from a byte slice.
func ParseTenc(data []byte) (TencBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TencBox{}, err
   }
   if len(data) < 8+1+3+2+1+16 {
      return TencBox{}, fmt.Errorf("tenc box is too small: %d bytes", len(data))
   }

   var tenc TencBox
   tenc.Header = header
   tenc.Version = data[8]
   copy(tenc.Flags[:], data[9:12])
   // Skip 2 bytes of reserved + isProtected
   tenc.DefaultPerSampleIVSize = data[14]
   copy(tenc.DefaultKID[:], data[15:31])

   return tenc, nil
}

// Encode encodes the 'tenc' box to a byte slice.
func (b *TencBox) Encode() []byte {
   // For round-trip, we assume the original data is stored if needed.
   // This simplified encode is for creating new boxes if required.
   encoded := make([]byte, 32) // size is fixed
   b.Header.Size = 32
   copy(b.Header.Type[:], "tenc")
   b.Header.Write(encoded)

   encoded[8] = b.Version
   copy(encoded[9:12], b.Flags[:])
   encoded[14] = b.DefaultPerSampleIVSize
   copy(encoded[15:31], b.DefaultKID[:])
   return encoded
}
