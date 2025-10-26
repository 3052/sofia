package mp4

// SinfBox represents the 'sinf' box.
type SinfBox struct {
   Header BoxHeader
   Frma   FrmaBox
   Schi   SchiBox
}

// ParseSinf parses the 'sinf' box from a byte slice.
func ParseSinf(data []byte) (SinfBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return SinfBox{}, err
   }
   var sinf SinfBox
   sinf.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return SinfBox{}, err
      }
      switch string(h.Type[:]) {
      case "frma":
         frma, err := ParseFrma(boxData[offset:])
         if err != nil {
            return SinfBox{}, err
         }
         sinf.Frma = frma
         offset += int(frma.Header.Size)
      case "schi":
         schi, err := ParseSchi(boxData[offset:])
         if err != nil {
            return SinfBox{}, err
         }
         sinf.Schi = schi
         offset += int(schi.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   return sinf, nil
}

// Encode encodes the 'sinf' box to a byte slice.
func (b *SinfBox) Encode() []byte {
   content := b.Frma.Encode()
   content = append(content, b.Schi.Encode()...)
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
