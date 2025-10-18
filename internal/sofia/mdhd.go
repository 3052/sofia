// File: mdhd_box.go
package mp4parser

// MdhdBox (Media Header Box)
type MdhdBox struct {
   FullBox
   RemainingData []byte
}

func ParseMdhdBox(data []byte) (*MdhdBox, error) {
   b := &MdhdBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   b.RemainingData = data[offset:]
   return b, nil
}
func (b *MdhdBox) Size() uint64 {
   return 8 + b.FullBox.Size() + uint64(len(b.RemainingData))
}
func (b *MdhdBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "mdhd")
   offset = b.FullBox.Format(dst, offset)
   offset = writeBytes(dst, offset, b.RemainingData)
   return offset
}
