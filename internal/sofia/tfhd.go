package mp4

// TfhdBox represents the 'tfhd' box.
type TfhdBox struct {
   Header BoxHeader
   Data   []byte
}

// ParseTfhd parses the 'tfhd' box from a byte slice.
func ParseTfhd(data []byte) (TfhdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TfhdBox{}, err
   }
   return TfhdBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'tfhd' box to a byte slice.
func (b *TfhdBox) Encode() []byte {
   return b.Data
}
