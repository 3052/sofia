package mp4

import (
   "encoding/binary"
   "errors"
)

// PsshBox represents the 'pssh' box (Protection System Specific Header).
// It now contains fully parsed fields.
type PsshBox struct {
   Header   BoxHeader
   Version  byte
   Flags    [3]byte
   SystemID [16]byte
   KIDs     [][16]byte
   Data     []byte
}

// ParsePssh now fully parses the pssh box structure.
func ParsePssh(data []byte) (PsshBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return PsshBox{}, err
   }
   var pssh PsshBox
   pssh.Header = header

   // A pssh is a Full Box
   if len(data) < 12 {
      return PsshBox{}, errors.New("pssh box is too short for version and flags")
   }
   pssh.Version = data[8]
   copy(pssh.Flags[:], data[9:12])
   offset := 12

   if len(data) < offset+16 {
      return PsshBox{}, errors.New("pssh box is too short for SystemID")
   }
   copy(pssh.SystemID[:], data[offset:offset+16])
   offset += 16

   if pssh.Version > 0 {
      if len(data) < offset+4 {
         return PsshBox{}, errors.New("pssh v1+ box is too short for KIDCount")
      }
      kidCount := binary.BigEndian.Uint32(data[offset : offset+4])
      offset += 4
      pssh.KIDs = make([][16]byte, kidCount)
      for i := 0; i < int(kidCount); i++ {
         if len(data) < offset+16 {
            return PsshBox{}, errors.New("pssh box is truncated while parsing KIDs")
         }
         var kid [16]byte
         copy(kid[:], data[offset:offset+16])
         pssh.KIDs[i] = kid
         offset += 16
      }
   }

   if len(data) < offset+4 {
      return PsshBox{}, errors.New("pssh box is too short for DataSize")
   }
   dataSize := binary.BigEndian.Uint32(data[offset : offset+4])
   offset += 4

   if len(data) < offset+int(dataSize) {
      return PsshBox{}, errors.New("pssh data size exceeds box bounds")
   }
   pssh.Data = data[offset : offset+int(dataSize)]

   return pssh, nil
}

// Encode now correctly serializes the box from its fields.
func (b *PsshBox) Encode() []byte {
   kidSectionSize := 0
   if b.Version > 0 {
      kidSectionSize = 4 + (len(b.KIDs) * 16)
   }
   dataSize := len(b.Data)

   // Calculate total size of the box payload
   payloadSize := 4 + 16 + kidSectionSize + 4 + dataSize // FullBoxHeader + SystemID + [KIDs] + DataSize + Data
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
