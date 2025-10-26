package mp4

// MdatBox represents the 'mdat' box.
type MdatBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseMdat parses the 'mdat' box from a byte slice.
func ParseMdat(data []byte) (MdatBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdatBox{}, err
   }
   return MdatBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'mdat' box to a byte slice.
func (b *MdatBox) Encode() []byte {
   return b.Data
}
