package mp4

// MdiaBox represents the 'mdia' box.
type MdiaBox struct {
   Header BoxHeader
   Mdhd   MdhdBox
   Minf   MinfBox
}

// ParseMdia parses the 'mdia' box from a byte slice.
func ParseMdia(data []byte) (MdiaBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MdiaBox{}, err
   }
   var mdia MdiaBox
   mdia.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MdiaBox{}, err
      }
      switch string(h.Type[:]) {
      case "mdhd":
         mdhd, err := ParseMdhd(boxData[offset:])
         if err != nil {
            return MdiaBox{}, err
         }
         mdia.Mdhd = mdhd
         offset += int(mdhd.Header.Size)
      case "minf":
         minf, err := ParseMinf(boxData[offset:])
         if err != nil {
            return MdiaBox{}, err
         }
         mdia.Minf = minf
         offset += int(minf.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   return mdia, nil
}

// Encode encodes the 'mdia' box to a byte slice.
func (b *MdiaBox) Encode() []byte {
   content := b.Mdhd.Encode()
   content = append(content, b.Minf.Encode()...)
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
