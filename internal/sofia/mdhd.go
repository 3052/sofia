package mp4

// MdhdBox represents the 'mdhd' box.
type MdhdBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseMdhd parses the 'mdhd' box from a byte slice.
func ParseMdhd(data []byte) (MdhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdhdBox{}, err
   }
   return MdhdBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'mdhd' box to a byte slice.
func (b *MdhdBox) Encode() []byte {
   return b.Data
}
