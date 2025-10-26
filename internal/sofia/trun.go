package mp4

// TrunBox represents the 'trun' box.
type TrunBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseTrun parses the 'trun' box from a byte slice.
func ParseTrun(data []byte) (TrunBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrunBox{}, err
   }
   return TrunBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'trun' box to a byte slice.
func (b *TrunBox) Encode() []byte {
   return b.Data
}
