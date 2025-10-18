package mp4parser

// EncaChildBox can hold any of the parsed child types or a raw box.
type EncaChildBox struct {
   Sinf *SinfBox
   Raw  *RawBox
}

// Size calculates the size of the contained child.
func (c *EncaChildBox) Size() uint64 {
   if c.Sinf != nil {
      return c.Sinf.Size()
   }
   if c.Raw != nil {
      return c.Raw.Size()
   }
   return 0
}

// Format formats the contained child.
func (c *EncaChildBox) Format(dst []byte, offset int) int {
   if c.Sinf != nil {
      return c.Sinf.Format(dst, offset)
   }
   if c.Raw != nil {
      return c.Raw.Format(dst, offset)
   }
   return offset
}

// EncaBox (Encrypted Audio Sample Entry)
type EncaBox struct {
   Type string
   // The 28 bytes of the AudioSampleEntry prefix.
   Prefix   []byte
   Children []*EncaChildBox
}

// ParseEncaBox parses the EncaBox from its content slice.
func ParseEncaBox(data []byte) (*EncaBox, error) {
   b := &EncaBox{}
   b.Type = "enca" // Default to the parsed type.

   // AudioSampleEntry has a 28-byte prefix before its child boxes.
   const prefixSize = 28
   if len(data) < prefixSize {
      return nil, ErrUnexpectedEOF
   }
   b.Prefix = data[:prefixSize]

   // Child boxes start AFTER the prefix.
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
      child := &EncaChildBox{}
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

// Size calculates the total byte size of the EncaBox.
func (b *EncaBox) Size() uint64 {
   size := uint64(8 + len(b.Prefix))
   for _, child := range b.Children {
      size += child.Size()
   }
   return size
}

// Format serializes the EncaBox into the destination slice using its Type field.
func (b *EncaBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, b.Type)
   offset = writeBytes(dst, offset, b.Prefix)
   for _, child := range b.Children {
      offset = child.Format(dst, offset)
   }
   return offset
}
