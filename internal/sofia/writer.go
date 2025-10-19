// File: binary_writer.go
package mp4parser

import "encoding/binary"

func writeUint8(dst []byte, offset int, val uint8) int {
   dst[offset] = val
   return offset + 1
}

func writeUint32(dst []byte, offset int, val uint32) int {
   binary.BigEndian.PutUint32(dst[offset:], val)
   return offset + 4
}

func writeUint64(dst []byte, offset int, val uint64) int {
   binary.BigEndian.PutUint64(dst[offset:], val)
   return offset + 8
}

func writeString(dst []byte, offset int, val string) int {
   copy(dst[offset:], val)
   return offset + len(val)
}

func writeBytes(dst []byte, offset int, data []byte) int {
   copy(dst[offset:], data)
   return offset + len(data)
}
