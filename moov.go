package sofia

import (
   "bytes"
   "errors"
)

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

func (b *MoovBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in moov")
      }
      childData := boxData[offset : offset+boxSize]
      var child MoovChild
      switch string(h.Type[:]) {
      case "trak":
         var trak TrakBox
         if err := trak.Parse(childData); err != nil {
            return err
         }
         child.Trak = &trak
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(childData); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = childData
      }
      b.Children = append(b.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return nil
}

func (b *MoovBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      // Logic update: Skip 'pssh' entirely to delete it.
      if child.Trak != nil {
         content = append(content, child.Trak.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

func (b *MoovBox) Trak() (*TrakBox, bool) {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak, true
      }
   }
   return nil, false
}

// FindPssh finds the first PsshBox child with a matching SystemID.
func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, child := range b.Children {
      if child.Pssh != nil {
         if bytes.Equal(child.Pssh.SystemID[:], systemID) {
            return child.Pssh, true
         }
      }
   }
   return nil, false
}
