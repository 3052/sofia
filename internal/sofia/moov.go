package mp4

import "fmt"

type MoovChild struct {
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

type MoovBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MoovChild
}

func ParseMoov(data []byte) (MoovBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return MoovBox{}, err
   }
   var moov MoovBox
   moov.Header = header
   moov.RawData = data[:header.Size]
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
         return MoovBox{}, fmt.Errorf("invalid child box size in moov")
      }
      childData := boxData[offset : offset+boxSize]
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
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return moov, nil
}

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

func (b *MoovBox) GetTrakByTrackID(trackID uint32) *TrakBox {
   for _, child := range b.Children {
      if child.Trak != nil {
         if trackID == 1 {
            return child.Trak
         }
      }
   }
   return nil
}

func (b *MoovBox) GetAllTraks() []*TrakBox {
   var traks []*TrakBox
   for _, child := range b.Children {
      if child.Trak != nil {
         traks = append(traks, child.Trak)
      }
   }
   return traks
}
