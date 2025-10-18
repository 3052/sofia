package mp4parser

type StblChildBox struct {
   Stsd *StsdBox
   Raw  *RawBox
}

func (c *StblChildBox) Size() uint64 {
   if c.Stsd != nil {
      return c.Stsd.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}
func (c *StblChildBox) Format(dst []byte, offset int) int {
   if c.Stsd != nil {
      return c.Stsd.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type StblBox struct{ Children []*StblChildBox }

func ParseStblBox(data []byte) (*StblBox, error) {
   b := &StblBox{}
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
      child := &StblChildBox{}
      switch header.Type {
      case "stsd":
         child.Stsd, err = ParseStsdBox(content)
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
func (b *StblBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}
func (b *StblBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "stbl")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
