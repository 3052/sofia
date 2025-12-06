package sofia

import "bytes"

type MoovChild struct {
   Mvhd *MvhdBox
   Trak *TrakBox
   Pssh *PsshBox
   Raw  []byte
}

type MoovBox struct {
   Header   BoxHeader
   Children []MoovChild
}

func (b *MoovBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MoovChild
      switch string(h.Type[:]) {
      case "mvhd":
         var mvhd MvhdBox
         if err := mvhd.Parse(content); err != nil {
            return err
         }
         child.Mvhd = &mvhd
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
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mvhd != nil {
         buf = append(buf, child.Mvhd.Encode()...)
      } else if child.Trak != nil {
         buf = append(buf, child.Trak.Encode()...)
      } else if child.Pssh != nil {
         // Skipped
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   b.Header.Put(buf)
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

func (b *MoovBox) Mvhd() (*MvhdBox, bool) {
   for _, child := range b.Children {
      if child.Mvhd != nil {
         return child.Mvhd, true
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
