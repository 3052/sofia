package mp4

// TrakBox represents the 'trak' box.
type TrakBox struct {
   Header BoxHeader
   Mdia   MdiaBox
}

// ParseTrak parses the 'trak' box from a byte slice.
func ParseTrak(data []byte) (TrakBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrakBox{}, err
   }
   var trak TrakBox
   trak.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return TrakBox{}, err
      }
      if string(h.Type[:]) == "mdia" {
         mdia, err := ParseMdia(boxData[offset:])
         if err != nil {
            return TrakBox{}, err
         }
         trak.Mdia = mdia
         offset += int(mdia.Header.Size)
      } else {
         offset += int(h.Size)
      }
   }
   return trak, nil
}

// Encode encodes the 'trak' box to a byte slice.
func (b *TrakBox) Encode() []byte {
   content := b.Mdia.Encode()
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
