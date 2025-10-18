package mp4parser

import "bytes"

// PsshBox (Protection System Specific Header Box)
type PsshBox struct {
   FullBox
   SystemID []byte // 16 bytes
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

   isWidevine := bytes.Equal(b.SystemID, []byte{0xed, 0xef, 0x8b, 0xa9, 0x79, 0xd6, 0x4a, 0xce, 0xa3, 0xc8, 0x27, 0xdc, 0xd5, 0x1d, 0x21, 0xed})

   if isWidevine {
      if b.Version > 0 {
         var kidCount uint32
         kidCount, offset, err = readUint32(data, offset)
         if err != nil {
            return nil, err
         }
         offset += int(kidCount * 16)
      }
      var dataSize uint32
      dataSize, offset, err = readUint32(data, offset)
      if err != nil {
         return nil, err
      }
      if offset+int(dataSize) > len(data) {
         return nil, ErrUnexpectedEOF
      }
      b.Data = data[offset : offset+int(dataSize)]
   } else {
      b.Data = data[offset:]
   }
   return b, nil
}
func (b *PsshBox) Size() uint64 {
   // This is tricky, the best way to roundtrip pssh is to just store its content
   // as raw bytes inside the struct. For now, let's keep the content.
   return 8 + b.FullBox.Size() + uint64(len(b.SystemID)) + uint64(len(b.Data))
}
func (b *PsshBox) Format(dst []byte, offset int) int {
   // A raw-byte approach would be better here for perfect roundtrip.
   // For now, this rebuilds a simplified version.
   contentLen := int(b.FullBox.Size() + uint64(len(b.SystemID)) + uint64(len(b.Data)))
   offset = writeUint32(dst, offset, uint32(8+contentLen))
   offset = writeString(dst, offset, "pssh")
   offset = b.FullBox.Format(dst, offset)
   offset = writeBytes(dst, offset, b.SystemID)
   offset = writeBytes(dst, offset, b.Data)
   return offset
}
