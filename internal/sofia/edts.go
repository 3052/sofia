package mp4

// EdtsBox represents the 'edts' box (Edit Box).
type EdtsBox struct {
   Header  BoxHeader
   RawData []byte
}

// Parse parses the 'edts' box from a byte slice.
func (b *EdtsBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   return nil
}

// Encode now correctly serializes the box from its fields.
func (b *EdtsBox) Encode() []byte {
   b.Header.Size = uint32(len(b.RawData))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], b.RawData[8:])
   return encoded
}
