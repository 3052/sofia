package mp4

// TrafBox represents the 'traf' box.
type TrafBox struct {
   Header BoxHeader
   Tfhd   TfhdBox
   Trun   *TrunBox
   Senc   *SencBox
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
      switch string(h.Type[:]) {
      case "tfhd":
         tfhd, err := ParseTfhd(boxData[offset:])
         if err != nil {
            return TrafBox{}, err
         }
         traf.Tfhd = tfhd
         offset += int(tfhd.Header.Size)
      case "trun":
         trun, err := ParseTrun(boxData[offset:])
         if err != nil {
            return TrafBox{}, err
         }
         traf.Trun = &trun
         offset += int(trun.Header.Size)
      case "senc":
         senc, err := ParseSenc(boxData[offset:])
         if err != nil {
            return TrafBox{}, err
         }
         traf.Senc = &senc
         offset += int(senc.Header.Size)
      default:
         offset += int(h.Size)
      }
   }
   return traf, nil
}

// Encode encodes the 'traf' box to a byte slice.
func (b *TrafBox) Encode() []byte {
   content := b.Tfhd.Encode()
   if b.Trun != nil {
      content = append(content, b.Trun.Encode()...)
   }
   if b.Senc != nil {
      content = append(content, b.Senc.Encode()...)
   }
   b.Header.Size = uint32(len(content) + 8)
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
