package mp4parser

type MoofChildBox struct {
   Traf *TrafBox
   Raw  *RawBox
}

func (c *MoofChildBox) Size() uint64 {
   if c.Traf != nil {
      return c.Traf.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}
func (c *MoofChildBox) Format(dst []byte, offset int) int {
   if c.Traf != nil {
      return c.Traf.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type MoofBox struct{ Children []*MoofChildBox }

func ParseMoofBox(data []byte) (*MoofBox, error) {
   b := &MoofBox{}
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
      child := &MoofChildBox{}
      switch header.Type {
      case "traf":
         child.Traf, err = ParseTrafBox(content)
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
func (b *MoofBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}
func (b *MoofBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "moof")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
