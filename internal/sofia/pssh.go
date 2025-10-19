// File: pssh_box.go
package mp4parser

type PsshBox struct {
   FullBox
   SystemID []byte
   Data     []byte
}

func ParsePsshBox(data []byte) (*PsshBox, error) {
   b := &PsshBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }
   if offset+16 > len(data) {
      return nil, ErrUnexpectedEOF
   }
   b.SystemID = data[offset : offset+16]
   offset += 16
   b.Data = data[offset:]
   return b, nil
}
func (b *PsshBox) Size() uint64 {
   return 8 + b.FullBox.Size() + 16 + uint64(len(b.Data))
}
func (b *PsshBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "pssh")
   offset = b.FullBox.Format(dst, offset)
   offset = writeBytes(dst, offset, b.SystemID)
   offset = writeBytes(dst, offset, b.Data)
   return offset
}
