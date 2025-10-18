// File: sinf_box.go
package mp4parser

type SinfChildBox struct {
   Frma *FrmaBox
   Schi *SchiBox
   Raw  *RawBox
}

func (c *SinfChildBox) Size() uint64 {
   switch {
   case c.Frma != nil:
      return c.Frma.Size()
   case c.Schi != nil:
      return c.Schi.Size()
   case c.Raw != nil:
      return c.Raw.Size()
   }
   return 0
}
func (c *SinfChildBox) Format(dst []byte, offset int) int {
   switch {
   case c.Frma != nil:
      return c.Frma.Format(dst, offset)
   case c.Schi != nil:
      return c.Schi.Format(dst, offset)
   case c.Raw != nil:
      return c.Raw.Format(dst, offset)
   }
   return offset
}

type SinfBox struct{ Children []*SinfChildBox }

func ParseSinfBox(data []byte) (*SinfBox, error) {
   b := &SinfBox{}
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
      child := &SinfChildBox{}
      switch header.Type {
      case "frma":
         child.Frma, err = ParseFrmaBox(content)
      case "schi":
         child.Schi, err = ParseSchiBox(content)
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
func (b *SinfBox) Size() uint64 {
   size := uint64(8)
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}
func (b *SinfBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "sinf")
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
