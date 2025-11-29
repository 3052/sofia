package sofia

import (
   "bytes"
   "encoding/binary"
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
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MoovChild
      switch string(h.Type[:]) {
      case "trak":
         var trak TrakBox
         if err := trak.Parse(content); err != nil {
            return err
         }
         child.Trak = &trak
      case "pssh":
         var pssh PsshBox
         if err := pssh.Parse(content); err != nil {
            return err
         }
         child.Pssh = &pssh
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MoovBox) Encode() []byte {
   // 1. Start with placeholder for header
   buf := make([]byte, 8)

   // 2. Append children
   for _, child := range b.Children {
      if child.Trak != nil {
         buf = append(buf, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         // Skipped (Read-only)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }

   // 3. Set Header
   b.Header.Size = uint32(len(buf))
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])

   return buf
}

func (b *MoovBox) RemovePssh() {
   var kept []MoovChild
   for _, child := range b.Children {
      if child.Pssh != nil {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) RemoveMvex() {
   var kept []MoovChild
   for _, child := range b.Children {
      if len(child.Raw) >= 8 && string(child.Raw[4:8]) == "mvex" {
         continue
      }
      kept = append(kept, child)
   }
   b.Children = kept
}

func (b *MoovBox) Trak() (*TrakBox, bool) {
   for _, child := range b.Children {
      if child.Trak != nil {
         return child.Trak, true
      }
   }
   return nil, false
}

func (b *MoovBox) FindPssh(systemID []byte) (*PsshBox, bool) {
   for _, child := range b.Children {
      if child.Pssh != nil && bytes.Equal(child.Pssh.SystemID[:], systemID) {
         return child.Pssh, true
      }
   }
   return nil, false
}
