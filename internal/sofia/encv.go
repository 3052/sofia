package mp4

// EncvBox represents the 'encv' box.
type EncvBox struct {
   Header BoxHeader
   Data   []byte
   Sinf   SinfBox
}

// ParseEncv parses the 'encv' box from a byte slice.
func ParseEncv(data []byte) (EncvBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return EncvBox{}, err
   }
   var encv EncvBox
   encv.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return EncvBox{}, err
      }
      switch string(h.Type[:]) {
      case "sinf":
         sinf, err := ParseSinf(boxData[offset:])
         if err != nil {
            return EncvBox{}, err
         }
         encv.Sinf = sinf
         offset += int(sinf.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   encv.Data = data[:header.Size]
   return encv, nil
}

// Encode encodes the 'encv' box to a byte slice.
func (b *EncvBox) Encode() []byte {
   return b.Data
}
