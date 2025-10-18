// File: schi_box.go
package mp4parser

type SchiChildBox struct {
   Tenc *TencBox
   Raw  *RawBox
}

func (c *SchiChildBox) Size() uint64 {
   if c.Tenc != nil {
      return c.Tenc.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}
func (c *SchiChildBox) Format(dst []byte, offset int) int {
   if c.Tenc != nil {
      return c.Tenc.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type SchiBox struct{ Children []*SchiChildBox }

func ParseSchiBox(data []byte) (*SchiBox, error) {
   b := &SchiBox{}
   offset := 0
   for offset < len(data) {
      header, headerEndOffset, err := ParseBoxHeader(data, offset)
      if err != nil {
         return nil, err
      }
      contentEndOffset := offset + int(header.Size)
      content := data[headerEndOffset:contentEndOffset]
      child := &SchiChildBox{}
      switch header.Type {
      case "tenc":
         child.Tenc, err = ParseTencBox(content)
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
func (b *SchiBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}
func (b *SchiBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "schi")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
