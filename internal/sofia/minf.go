package mp4

// MinfBox represents the 'minf' box.
type MinfBox struct {
   Header BoxHeader
   Stbl   StblBox
}

// ParseMinf parses the 'minf' box from a byte slice.
func ParseMinf(data []byte) (MinfBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MinfBox{}, err
   }
   var minf MinfBox
   minf.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MinfBox{}, err
      }
      if string(h.Type[:]) == "stbl" {
         stbl, err := ParseStbl(boxData[offset:])
         if err != nil {
            return MinfBox{}, err
         }
         minf.Stbl = stbl
         offset += int(stbl.Header.Size)
      } else {
         offset += int(h.Size)
      }
   }
   return minf, nil
}

// Encode encodes the 'minf' box to a byte slice.
func (b *MinfBox) Encode() []byte {
   content := b.Stbl.Encode()
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
