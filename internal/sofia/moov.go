// File: moov_box.go
package mp4parser

type MoovChildBox struct {
   Trak *TrakBox
   Pssh *PsshBox
   Raw  *RawBox
}

func (c *MoovChildBox) Size() uint64 {
   switch {
   case c.Trak != nil:
      return c.Trak.Size()
   case c.Pssh != nil:
      return c.Pssh.Size()
   case c.Raw != nil:
      return c.Raw.Size()
   }
   return 0
}
func (c *MoovChildBox) Format(dst []byte, offset int) int {
   switch {
   case c.Trak != nil:
      return c.Trak.Format(dst, offset)
   case c.Pssh != nil:
      return c.Pssh.Format(dst, offset)
   case c.Raw != nil:
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type MoovBox struct{ Children []*MoovChildBox }

func ParseMoovBox(data []byte) (*MoovBox, error) {
   b := &MoovBox{}
   offset := 0
   for offset < len(data) {
      header, headerEndOffset, err := ParseBoxHeader(data, offset)
      if err != nil {
         return nil, err
      }
      contentEndOffset := offset + int(header.Size)
      if contentEndOffset > len(data) {
         return nil, ErrUnexpectedEOF
      }
      content := data[headerEndOffset:contentEndOffset]
      child := &MoovChildBox{}
      switch header.Type {
      case "trak":
         child.Trak, err = ParseTrakBox(content)
      case "pssh":
         child.Pssh, err = ParsePsshBox(content)
      default:
         child.Raw, err = ParseRawBox(header.Type, content)
      }
      if err != nil {
         return nil, err
      }
      b.Children = append(b.Children, child)
      offset = contentEndOffset
   }
   return b, nil
}
func (b *MoovBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}
func (b *MoovBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "moov")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
