package mp4

// PsshBox represents the 'pssh' box.
type PsshBox struct {
   Header BoxHeader
   Data   []byte
}

// ParsePssh parses the 'pssh' box from a byte slice.
func ParsePssh(data []byte) (PsshBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return PsshBox{}, err
   }
   return PsshBox{
      Header: header,
      Data:   data[:header.Size],
   }, nil
}

// Encode encodes the 'pssh' box to a byte slice.
func (b *PsshBox) Encode() []byte {
   return b.Data
}
