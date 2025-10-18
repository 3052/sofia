// File: encv_box.go
package mp4parser

// EncvBox (Encrypted Video Sample Entry)
type EncvBox struct {
   // The 78 bytes of the VisualSampleEntry prefix.
   Prefix   []byte
   Children []*EncvChildBox
}

type EncvChildBox struct {
   Sinf *SinfBox
   Raw  *RawBox
}

func (c *EncvChildBox) Size() uint64 {
   if c.Sinf != nil {
      return c.Sinf.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}
func (c *EncvChildBox) Format(dst []byte, offset int) int {
   if c.Sinf != nil {
      return c.Sinf.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

func ParseEncvBox(data []byte) (*EncvBox, error) {
   b := &EncvBox{}
   const prefixSize = 78
   if len(data) < prefixSize {
      return nil, ErrUnexpectedEOF
   }
   b.Prefix = data[:prefixSize]

   offset := prefixSize
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
      child := &EncvChildBox{}
      switch header.Type {
      case "sinf":
         child.Sinf, err = ParseSinfBox(content)
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

func (b *EncvBox) Size() uint64 {
   size := uint64(8 + len(b.Prefix))
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}

func (b *EncvBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "encv")
   offset = writeBytes(dst, offset, b.Prefix)
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
