package mp4

// FrmaBox represents the 'frma' box.
type FrmaBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseFrma parses the 'frma' box from a byte slice.
func ParseFrma(data []byte) (FrmaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return FrmaBox{}, err
   }
   return FrmaBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'frma' box to a byte slice.
func (b *FrmaBox) Encode() []byte {
   return b.Data
}
