package sofia

import (
   "bytes"
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

func (b Box) Type() string {
   return string(b.Header.Type[:])
}

func (b Box) Boxes() ([]Box, error) {
   r := bytes.NewReader(b.Payload)
   var vs []Box
   for {
      var v Box
      err := v.Decode(r)
      if err == io.EOF {
         return vs, nil
      } else if err != nil {
         return nil, err
      }
      vs = append(vs, v)
   }
}

func (b *Box) Decode(r io.Reader) error {
   err := binary.Read(r, binary.BigEndian, &b.Header.Size)
   if err != nil {
      return err
   }
   _, err = r.Read(b.Header.Type[:])
   if err != nil {
      return err
   }
   b.Payload = make([]byte, b.Header.Size)
   _, err = r.Read(b.Payload)
   if err != nil {
      return err
   }
   return nil
}
