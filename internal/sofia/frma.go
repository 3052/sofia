// File: frma_box.go
package mp4parser

type FrmaBox struct{ DataFormat []byte }

func ParseFrmaBox(data []byte) (*FrmaBox, error) {
   if len(data) < 4 {
      return nil, ErrUnexpectedEOF
   }
   return &FrmaBox{DataFormat: data[:4]}, nil
}
func (b *FrmaBox) Size() uint64 { return 8 + 4 }
func (b *FrmaBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "frma")
   offset = writeBytes(dst, offset, b.DataFormat)
   return offset
}
