// File: sidx_box.go
package mp4parser

type SidxBox struct {
   FullBox
   RemainingData []byte
}

func ParseSidxBox(data []byte) (*SidxBox, error) {
   b := &SidxBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   b.RemainingData = data[offset:]
   return b, nil
}
func (b *SidxBox) Size() uint64 { return 8 + b.FullBox.Size() + uint64(len(b.RemainingData)) }
func (b *SidxBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "sidx")
   offset = b.FullBox.Format(dst, offset)
   offset = writeBytes(dst, offset, b.RemainingData)
   return offset
}
