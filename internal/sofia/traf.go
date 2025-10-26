package mp4

// TrafChild holds either a parsed box or raw data for a child of a 'traf' box.
type TrafChild struct {
   Tfhd *TfhdBox
   Trun *TrunBox
   Senc *SencBox
   Raw  []byte
}

// TrafBox represents the 'traf' box (Track Fragment Box).
type TrafBox struct {
   Header   BoxHeader
   Children []TrafChild
}

// ParseTraf parses the 'traf' box from a byte slice.
func ParseTraf(data []byte) (TrafBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrafBox{}, err
   }
   var traf TrafBox
   traf.Header = header
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         return TrafBox{}, err
      }

      childData := boxData[offset : offset+int(h.Size)]
      var child TrafChild

      switch string(h.Type[:]) {
      case "tfhd":
         tfhd, err := ParseTfhd(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Tfhd = &tfhd
      case "trun":
         trun, err := ParseTrun(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Trun = &trun
      case "senc":
         senc, err := ParseSenc(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Senc = &senc
      default:
         child.Raw = childData
      }
      traf.Children = append(traf.Children, child)
      offset += int(h.Size)
   }
   return traf, nil
}

// Encode encodes the 'traf' box to a byte slice.
func (b *TrafBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Tfhd != nil {
         content = append(content, child.Tfhd.Encode()...)
      } else if child.Trun != nil {
         content = append(content, child.Trun.Encode()...)
      } else if child.Senc != nil {
         content = append(content, child.Senc.Encode()...)
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
