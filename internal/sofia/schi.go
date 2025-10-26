package mp4

// SchiBox represents the 'schi' box.
type SchiBox struct {
   Header BoxHeader
   Tenc   TencBox
}

// ParseSchi parses the 'schi' box from a byte slice.
func ParseSchi(data []byte) (SchiBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SchiBox{}, err
   }
   var schi SchiBox
   schi.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return SchiBox{}, err
      }
      if string(h.Type[:]) == "tenc" {
         tenc, err := ParseTenc(boxData[offset:])
         if err != nil {
            return SchiBox{}, err
         }
         schi.Tenc = tenc
         offset += int(tenc.Header.Size)
      } else {
         offset += int(h.Size)
      }
   }
   return schi, nil
}

// Encode encodes the 'schi' box to a byte slice.
func (b *SchiBox) Encode() []byte {
   content := b.Tenc.Encode()
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
