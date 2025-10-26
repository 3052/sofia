package mp4

// MoofBox represents the 'moof' box.
type MoofBox struct {
   Header BoxHeader
   Trafs  []TrafBox
}

// ParseMoof parses the 'moof' box from a byte slice.
func ParseMoof(data []byte) (MoofBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoofBox{}, err
   }
   var moof MoofBox
   moof.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MoofBox{}, err
      }
      if string(h.Type[:]) == "traf" {
         traf, err := ParseTraf(boxData[offset:])
         if err != nil {
            return MoofBox{}, err
         }
         moof.Trafs = append(moof.Trafs, traf)
         offset += int(traf.Header.Size)
      } else {
         offset += int(h.Size)
      }
   }
   return moof, nil
}

// Encode encodes the 'moof' box to a byte slice.
func (b *MoofBox) Encode() []byte {
   var content []byte
   for _, traf := range b.Trafs {
      content = append(content, traf.Encode()...)
   }
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
