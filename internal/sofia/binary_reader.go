// File: binary_reader.go
package mp4parser

import (
   "encoding/binary"
   "errors"
)

var ErrUnexpectedEOF = errors.New("unexpected end of data")

func readUint32(data []byte, offset int) (uint32, int, error) {
   if offset+4 > len(data) {
      return 0, offset, ErrUnexpectedEOF
   }
   val := binary.BigEndian.Uint32(data[offset:])
   return val, offset + 4, nil
}

func readUint64(data []byte, offset int) (uint64, int, error) {
   if offset+8 > len(data) {
      return 0, offset, ErrUnexpectedEOF
   }
   val := binary.BigEndian.Uint64(data[offset:])
   return val, offset + 8, nil
}

func readString(data []byte, offset int, size int) (string, int, error) {
   if offset+size > len(data) {
      return "", offset, ErrUnexpectedEOF
   }
   val := string(data[offset : offset+size])
   return val, offset + size, nil
}
