package mp4

// SencBox represents the 'senc' box.
type SencBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseSenc parses the 'senc' box from a byte slice.
func ParseSenc(data []byte) (SencBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SencBox{}, err
   }
   return SencBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'senc' box to a byte slice.
func (b *SencBox) Encode() []byte {
   return b.Data
}
