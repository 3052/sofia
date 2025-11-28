package sofia

import (
   "encoding/binary"
   "errors"
)

type PsshBox struct {
   Header   BoxHeader
   Version  byte
   Flags    [3]byte
   SystemID [16]byte
   KIDs     [][16]byte
   Data     []byte
}

func (b *PsshBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   if len(data) < 12 {
      return errors.New("pssh box is too short for version and flags")
   }
   b.Version = data[8]
   copy(b.Flags[:], data[9:12])
   offset := 12
   if len(data) < offset+16 {
      return errors.New("pssh box is too short for SystemID")
   }
   copy(b.SystemID[:], data[offset:offset+16])
   offset += 16
   if b.Version > 0 {
      if len(data) < offset+4 {
         return errors.New("pssh v1+ box is too short for KIDCount")
      }
      kidCount := binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      b.KIDs = make([][16]byte, kidCount)
      for i := 0; i < int(kidCount); i++ {
         if len(data) < offset+16 {
            return errors.New("pssh box is truncated while parsing KIDs")
         }
         var kid [16]byte
         copy(kid[:], data[offset:offset+16])
         b.KIDs[i] = kid
         offset += 16
      }
   }
   if len(data) < offset+4 {
      return errors.New("pssh box is too short for DataSize")
   }
   dataSize := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4
   if len(data) < offset+int(dataSize) {
      return errors.New("pssh data size exceeds box bounds")
   }
   b.Data = data[offset : offset+int(dataSize)]
   return nil
}
