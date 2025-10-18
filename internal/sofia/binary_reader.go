package mp4parser

import (
   "encoding/binary"
   "errors"
)

// ErrUnexpectedEOF is returned when a read operation goes past the end of the byte slice.
var ErrUnexpectedEOF = errors.New("unexpected end of data")

// readUint16 reads a 2-byte big-endian unsigned integer from the slice at the given offset.
func readUint16(data []byte, offset int) (uint16, int, error) {
   if offset+2 > len(data) {
      return 0, offset, ErrUnexpectedEOF
   }
   val := binary.BigEndian.Uint16(data[offset:])
   return val, offset + 2, nil
}

// readUint32 reads a 4-byte big-endian unsigned integer from the slice at the given offset.
func readUint32(data []byte, offset int) (uint32, int, error) {
   if offset+4 > len(data) {
      return 0, offset, ErrUnexpectedEOF
   }
   val := binary.BigEndian.Uint32(data[offset:])
   return val, offset + 4, nil
}

// readUint64 reads an 8-byte big-endian unsigned integer from the slice at the given offset.
func readUint64(data []byte, offset int) (uint64, int, error) {
   if offset+8 > len(data) {
      return 0, offset, ErrUnexpectedEOF
   }
   val := binary.BigEndian.Uint64(data[offset:])
   return val, offset + 8, nil
}

// readString reads a string of a given size from the slice at the given offset.
func readString(data []byte, offset int, size int) (string, int, error) {
   if offset+size > len(data) {
      return "", offset, ErrUnexpectedEOF
   }
   val := string(data[offset : offset+size])
   return val, offset + size, nil
}
