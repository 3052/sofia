package mp4

// MoovChild holds either a parsed box or raw data for a child of a 'moov' box.
type MoovChild struct {
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

// MoovBox represents the 'moov' box (Movie Box).
type MoovBox struct {
   Header   BoxHeader
   Children []MoovChild
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

      childData := boxData[offset : offset+int(h.Size)]
      var child MoovChild

      switch string(h.Type[:]) {
      case "trak":
         trak, err := ParseTrak(childData)
         if err != nil {
            return MoovBox{}, err
         }
         child.Trak = &trak
      case "pssh":
         pssh, err := ParsePssh(childData)
         if err != nil {
            return MoovBox{}, err
         }
         child.Pssh = &pssh
      default:
         child.Raw = childData
      }
      moov.Children = append(moov.Children, child)
      offset += int(h.Size)
   }
   return moov, nil
}

// Encode encodes the 'moov' box to a byte slice.
func (b *MoovBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Trak != nil {
         content = append(content, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         content = append(content, child.Pssh.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }

   b.Header.Size = uint32(8 + len(content))
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
