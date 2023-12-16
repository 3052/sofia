package sofia

import (
   "encoding/binary"
   "io"
)

// aligned(8) class BoxHeader (
// unsigned int(32) boxtype,
// optional unsigned int(8)[16] extended_type) {
//    unsigned int(32) size;
//    unsigned int(32) type = boxtype;
//    if (size==1) {
//       unsigned int(64) largesize;
//    } else if (size==0) {
//       // box extends to end of file
//    }
//    if (boxtype=='uuid') {
//       unsigned int(8)[16] usertype = extended_type;
//    }
// }
type BoxHeader struct {
   Size uint32
   Type [4]byte
}

// aligned(8) class Box (
// unsigned int(32) boxtype,
// optional unsigned int(8)[16] extended_type) {
//    BoxHeader(boxtype, extended_type);
//    // the remaining bytes are the BoxPayload
// }
type Box struct {
   Header BoxHeader
   Payload []byte
}
