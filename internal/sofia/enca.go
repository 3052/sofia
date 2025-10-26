package mp4

// EncaBox represents the 'enca' box.
type EncaBox struct {
   Header BoxHeader
   Data   []byte
   Sinf   SinfBox
}

// ParseEnca parses the 'enca' box from a byte slice.
func ParseEnca(data []byte) (EncaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncaBox{}, err
   }
   var enca EncaBox
   enca.Header = header
   boxData := data[8:header.Size]

   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return EncaBox{}, err
      }
      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(boxData[offset:])
         if err != nil {
            return EncaBox{}, err
         }
         enca.Sinf = sinf
         offset += int(sinf.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   enca.Data = data[:header.Size]
   return enca, nil
}

// Encode encodes the 'enca' box to a byte slice.
func (b *EncaBox) Encode() []byte {
   return b.Data
}
