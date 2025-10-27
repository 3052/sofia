package mp4

// PsshBox represents the 'pssh' box (Protection System Specific Header).
type PsshBox struct {
   Header  BoxHeader
   RawData []byte
}

// ParsePssh parses the 'pssh' box from a byte slice.
func ParsePssh(data []byte) (PsshBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return PsshBox{}, err
   }
   var pssh PsshBox
   pssh.Header = header
   pssh.RawData = data[:header.Size]
   return pssh, nil
}

// Encode now correctly serializes the box from its fields.
func (b *PsshBox) Encode() []byte {
   // Reconstruct the box to honor any header modifications (like renaming to 'free').
   // The original payload (RawData[8:]) is preserved.
   b.Header.Size = uint32(len(b.RawData))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], b.RawData[8:])
   return encoded
}
