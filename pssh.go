package sofia

import (
   "bytes"
   "encoding/binary"
   "errors"
)

// FindPsshBySystemID finds the first PsshBox in a slice with a matching
// SystemID. It returns the box if found, otherwise nil.
func FindPsshBySystemID(psshBoxes []*PsshBox, systemID []byte) *PsshBox {
   // THIS IS THE CORRECTED LINE
   for _, pssh := range psshBoxes {
      if bytes.Equal(pssh.SystemID[:], systemID) {
         return pssh
      }
   }
   return nil
}

// PsshBox represents the 'pssh' box (Protection System Specific Header).
type PsshBox struct {
   Header   BoxHeader
   Version  byte
   Flags    [3]byte
   SystemID [16]byte
   KIDs     [][16]byte
   Data     []byte
}

// Parse now fully parses the pssh box structure.
func (b *PsshBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
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

// Encode now correctly serializes the box from its fields.
func (b *PsshBox) Encode() []byte {
   kidSectionSize := 0
   if b.Version > 0 {
      kidSectionSize = 4 + (len(b.KIDs) * 16)
   }
   dataSize := len(b.Data)

   payloadSize := 4 + 16 + kidSectionSize + 4 + dataSize
   b.Header.Size = uint32(8 + payloadSize)

   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)

   offset := 8
   encoded[offset] = b.Version
   copy(encoded[offset+1:offset+4], b.Flags[:])
   offset += 4

   copy(encoded[offset:offset+16], b.SystemID[:])
   offset += 16

   if b.Version > 0 {
      binary.BigEndian.PutUint32(encoded[offset:offset+4], uint32(len(b.KIDs)))
      offset += 4
      for _, kid := range b.KIDs {
         copy(encoded[offset:offset+16], kid[:])
         offset += 16
      }
   }

   binary.BigEndian.PutUint32(encoded[offset:offset+4], uint32(dataSize))
   offset += 4
   copy(encoded[offset:], b.Data)

   return encoded
}
