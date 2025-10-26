package mp4

// SidxBox represents the 'sidx' box.
type SidxBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseSidx parses the 'sidx' box from a byte slice.
func ParseSidx(data []byte) (SidxBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SidxBox{}, err
   }
   return SidxBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'sidx' box to a byte slice.
func (b *SidxBox) Encode() []byte {
   return b.Data
}
