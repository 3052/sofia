package mp4

// StblBox represents the 'stbl' box.
type StblBox struct {
   Header BoxHeader
   Stsd   StsdBox
}

// ParseStbl parses the 'stbl' box from a byte slice.
func ParseStbl(data []byte) (StblBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return StblBox{}, err
   }
   var stbl StblBox
   stbl.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return StblBox{}, err
      }
      if string(h.Type[:]) == "stsd" {
         stsd, err := ParseStsd(boxData[offset:])
         if err != nil {
            return StblBox{}, err
         }
         stbl.Stsd = stsd
         offset += int(stsd.Header.Size)
      } else {
         offset += int(h.Size)
      }
   }
   return stbl, nil
}

// Encode encodes the 'stbl' box to a byte slice.
func (b *StblBox) Encode() []byte {
   content := b.Stsd.Encode()
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
