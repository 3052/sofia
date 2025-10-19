// File: tenc_box.go
package mp4parser

import "log" // Import the log package

// TencBox (Track Encryption Box)
type TencBox struct {
   FullBox
   DefaultCryptByteBlock  uint8
   DefaultSkipByteBlock   uint8
   DefaultIsProtected     uint8
   DefaultPerSampleIVSize uint8
   DefaultKID             []byte // 16 bytes
   DefaultConstantIV      []byte
   // TrailingData captures any padding or unknown fields at the end for a perfect roundtrip.
   TrailingData []byte
}

// ParseTencBox parses the TencBox from its content slice.
func ParseTencBox(data []byte) (*TencBox, error) {
   b := &TencBox{}
   offset, err := b.FullBox.Parse(data, 0)
   if err != nil {
      return nil, err
   }

   log.Printf("[ParseTencBox] Parsing 'tenc' box with Version: %d", b.Version)

   if b.Version == 0 {
      offset += 2 // Skip 2 reserved bytes
      b.DefaultIsProtected, offset, err = readUint8(data, offset)
      if err != nil {
         return nil, err
      }
      b.DefaultPerSampleIVSize, offset, err = readUint8(data, offset)
      if err != nil {
         return nil, err
      }
   } else { // Version 1 or greater
      // CORRECTED LOGIC: Added skip for the first reserved byte in v1 boxes.
      offset++ // Skip 1 reserved byte

      var cryptSkip byte
      cryptSkip, offset, err = readUint8(data, offset)
      if err != nil {
         return nil, err
      }
      b.DefaultCryptByteBlock = (cryptSkip & 0xF0) >> 4
      b.DefaultSkipByteBlock = cryptSkip & 0x0F

      b.DefaultIsProtected, offset, err = readUint8(data, offset)
      if err != nil {
         return nil, err
      }
      b.DefaultPerSampleIVSize, offset, err = readUint8(data, offset)
      if err != nil {
         return nil, err
      }
   }

   // LOGGING: Log the crucial parsed values to confirm the fix
   log.Printf("[ParseTencBox] Parsed DefaultIsProtected: %d", b.DefaultIsProtected)
   log.Printf("[ParseTencBox] Parsed DefaultPerSampleIVSize: %d", b.DefaultPerSampleIVSize)

   if offset+16 > len(data) {
      return nil, ErrUnexpectedEOF
   }
   b.DefaultKID = data[offset : offset+16]
   offset += 16

   if b.DefaultIsProtected == 1 && b.DefaultPerSampleIVSize == 0 {
      if offset < len(data) {
         var constIVSize uint8
         constIVSize, offset, err = readUint8(data, offset)
         if err != nil {
            return nil, err
         }
         if offset+int(constIVSize) > len(data) {
            return nil, ErrUnexpectedEOF
         }
         b.DefaultConstantIV = data[offset : offset+int(constIVSize)]
         offset += int(constIVSize)
      }
   }

   if offset < len(data) {
      b.TrailingData = data[offset:]
   }
   log.Printf("[ParseTencBox] Successfully parsed 'tenc' box.")
   return b, nil
}

// Size calculates the total byte size of the TencBox.
func (b *TencBox) Size() uint64 {
   size := uint64(8) + b.FullBox.Size()
   if b.Version == 0 {
      size += 2 + 1 + 1 // reserved, protected, iv_size
   } else {
      // CORRECTED SIZE: Account for the added reserved byte
      size += 1 + 1 + 1 + 1 // reserved, crypt/skip, protected, iv_size
   }
   size += 16 // KID
   if b.DefaultIsProtected == 1 && b.DefaultPerSampleIVSize == 0 {
      size += 1 + uint64(len(b.DefaultConstantIV))
   }
   size += uint64(len(b.TrailingData))
   return size
}

// Format serializes the TencBox into the destination slice.
func (b *TencBox) Format(dst []byte, offset int) int {
   offset = writeUint32(dst, offset, uint32(b.Size()))
   offset = writeString(dst, offset, "tenc")
   offset = b.FullBox.Format(dst, offset)
   if b.Version == 0 {
      offset = writeUint16(dst, offset, 0)
      offset = writeUint8(dst, offset, b.DefaultIsProtected)
      offset = writeUint8(dst, offset, b.DefaultPerSampleIVSize)
   } else {
      // CORRECTED FORMAT: Write the reserved byte
      offset = writeUint8(dst, offset, 0) // reserved byte
      cryptSkip := (b.DefaultCryptByteBlock << 4) | b.DefaultSkipByteBlock
      offset = writeUint8(dst, offset, cryptSkip)
      offset = writeUint8(dst, offset, b.DefaultIsProtected)
      offset = writeUint8(dst, offset, b.DefaultPerSampleIVSize)
   }
   offset = writeBytes(dst, offset, b.DefaultKID)
   if b.DefaultIsProtected == 1 && b.DefaultPerSampleIVSize == 0 {
      offset = writeUint8(dst, offset, uint8(len(b.DefaultConstantIV)))
      offset = writeBytes(dst, offset, b.DefaultConstantIV)
   }
   if len(b.TrailingData) > 0 {
      offset = writeBytes(dst, offset, b.TrailingData)
   }
   return offset
}
