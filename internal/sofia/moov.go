package mp4

// MoovBox represents the 'moov' box.
type MoovBox struct {
   Header BoxHeader
   Traks  []TrakBox
   Pssh   []PsshBox
}

// ParseMoov parses the 'moov' box from a byte slice.
func ParseMoov(data []byte) (MoovBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoovBox{}, err
   }
   var moov MoovBox
   moov.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return MoovBox{}, err
      }
      switch string(h.Type[:]) {
      case "trak":
         trak, err := ParseTrak(boxData[offset:])
         if err != nil {
            return MoovBox{}, err
         }
         moov.Traks = append(moov.Traks, trak)
         offset += int(trak.Header.Size)
      case "pssh":
         pssh, err := ParsePssh(boxData[offset:])
         if err != nil {
            return MoovBox{}, err
         }
         moov.Pssh = append(moov.Pssh, pssh)
         offset += int(pssh.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   return moov, nil
}

// Encode encodes the 'moov' box to a byte slice.
func (b *MoovBox) Encode() []byte {
   var content []byte
   for _, trak := range b.Traks {
      content = append(content, trak.Encode()...)
   }
   for _, pssh := range b.Pssh {
      content = append(content, pssh.Encode()...)
   }
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
