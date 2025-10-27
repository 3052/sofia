package mp4

import "errors"

type MoofChild struct {
   Traf *TrafBox
   Pssh *PsshBox
   Raw  []byte
}

type MoofBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MoofChild
}

func ParseMoof(data []byte) (MoofBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoofBox{}, err
   }
   var moof MoofBox
   moof.Header = header
   moof.RawData = data[:header.Size]
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return MoofBox{}, errors.New("invalid child box size in moof")
      }
      childData := boxData[offset : offset+boxSize]
      var child MoofChild
      switch string(h.Type[:]) {
      case "traf":
         traf, err := ParseTraf(childData)
         if err != nil {
            return MoofBox{}, err
         }
         child.Traf = &traf
      case "pssh":
         pssh, err := ParsePssh(childData)
         if err != nil {
            return MoofBox{}, err
         }
         child.Pssh = &pssh
      default:
         child.Raw = childData
      }
      moof.Children = append(moof.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return moof, nil
}

func (b *MoofBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Traf != nil {
         content = append(content, child.Traf.Encode()...)
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
