package mp4parser

type TrakChildBox struct {
   Mdia *MdiaBox
   Raw  *RawBox
}

func (c *TrakChildBox) Size() uint64 {
   if c.Mdia != nil {
      return c.Mdia.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}

func (c *TrakChildBox) Format(dst []byte, offset int) int {
   if c.Mdia != nil {
      return c.Mdia.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type TrakBox struct{ Children []*TrakChildBox }

func ParseTrakBox(data []byte) (*TrakBox, error) {
   b := &TrakBox{}
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
      child := &TrakChildBox{}
      switch header.Type {
      case "mdia":
         child.Mdia, err = ParseMdiaBox(content)
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

func (b *TrakBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}

func (b *TrakBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "trak")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
