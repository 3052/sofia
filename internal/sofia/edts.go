package mp4

// EdtsBox represents the 'edts' box (Edit Box).
type EdtsBox struct {
   Header  BoxHeader
   RawData []byte
}

// ParseEdts parses the 'edts' box from a byte slice.
func ParseEdts(data []byte) (EdtsBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EdtsBox{}, err
   }
   var edts EdtsBox
   edts.Header = header
   edts.RawData = data[:header.Size]
   return edts, nil
}

// Encode now correctly serializes the box from its fields.
func (b *EdtsBox) Encode() []byte {
   b.Header.Size = uint32(len(b.RawData))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], b.RawData[8:])
   return encoded
}
