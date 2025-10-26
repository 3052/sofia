package mp4

import (
   "encoding/binary"
   "errors"
)

type BoxHeader struct {
   Size uint32
   Type [4]byte
}

func ReadBoxHeader(data []byte) (BoxHeader, int, error) {
   if len(data) < 8 {
      return BoxHeader{}, 0, errors.New("not enough data for box header")
   }
   var h BoxHeader
   h.Size = binary.BigEndian.Uint32(data[0:4])
   copy(h.Type[:], data[4:8])
   return h, 8, nil
}

func (h BoxHeader) Write(data []byte) int {
   binary.BigEndian.PutUint32(data[0:4], h.Size)
   copy(data[4:8], h.Type[:])
   return 8
}
