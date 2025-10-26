package mp4

// TencBox represents the 'tenc' box.
type TencBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseTenc parses the 'tenc' box from a byte slice.
func ParseTenc(data []byte) (TencBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TencBox{}, err
   }
   return TencBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'tenc' box to a byte slice.
func (b *TencBox) Encode() []byte {
   return b.Data
}
