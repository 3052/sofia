package mp4

// StsdBox represents the 'stsd' box.
type StsdBox struct {
   Header BoxHeader
   Encv   *EncvBox
   Enca   *EncaBox
}

// ParseStsd parses the 'stsd' box from a byte slice.
func ParseStsd(data []byte) (StsdBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StsdBox{}, err
   }
   var stsd StsdBox
   stsd.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return StsdBox{}, err
      }
      switch string(h.Type[:]) {
      case "encv":
         encv, err := ParseEncv(boxData[offset:])
         if err != nil {
            return StsdBox{}, err
         }
         stsd.Encv = &encv
         offset += int(encv.Header.Size)
      case "enca":
         enca, err := ParseEnca(boxData[offset:])
         if err != nil {
            return StsdBox{}, err
         }
         stsd.Enca = &enca
         offset += int(enca.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   return stsd, nil
}

// Encode encodes the 'stsd' box to a byte slice.
func (b *StsdBox) Encode() []byte {
   var content []byte
   if b.Encv != nil {
      content = append(content, b.Encv.Encode()...)
   }
   if b.Enca != nil {
      content = append(content, b.Enca.Encode()...)
   }
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
