package mp4

// EdtsBox represents the 'edts' box (Edit Box).
// We treat it as a leaf box since we only need to identify and rename it.
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

// Encode returns the raw byte data to ensure a perfect round trip.
func (b *EdtsBox) Encode() []byte {
   return b.RawData
}
