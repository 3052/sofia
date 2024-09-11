package pssh

import (
   "154.pages.dev/sofia"
   "encoding/binary"
   "io"
)

// ISO/IEC 23001-7
//   aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
//      'pssh', version, flags=0,
//   ) {
//      unsigned int(8)[16] SystemID;
//      if (version > 0) {
//         unsigned int(32) KID_count;
//         {
//            unsigned int(8)[16] KID;
//         } [KID_count];
//      }
//      unsigned int(32) DataSize;
//      unsigned int(8)[DataSize] Data;
//   }
type Box struct {
   BoxHeader     sofia.BoxHeader
   FullBoxHeader sofia.FullBoxHeader
   SystemId      sofia.UUID
   KidCount      uint32
   Kid           []sofia.UUID
   DataSize      uint32
   Data          []uint8
}

func (b *Box) Read(src io.Reader) error {
   err := b.FullBoxHeader.Read(src)
   if err != nil {
      return err
   }
   _, err = io.ReadFull(src, b.SystemId[:])
   if err != nil {
      return err
   }
   if b.FullBoxHeader.Version > 0 {
      err := binary.Read(src, binary.BigEndian, &b.KidCount)
      if err != nil {
         return err
      }
      b.Kid = make([]sofia.UUID, b.KidCount)
      err = binary.Read(src, binary.BigEndian, b.Kid)
      if err != nil {
         return err
      }
   }
   err = binary.Read(src, binary.BigEndian, &b.DataSize)
   if err != nil {
      return err
   }
   b.Data = make([]uint8, b.DataSize)
   _, err = io.ReadFull(src, b.Data)
   if err != nil {
      return err
   }
   return nil
}

func (b *Box) Write(dst io.Writer) error {
   err := b.BoxHeader.Write(dst)
   if err != nil {
      return err
   }
   err = b.FullBoxHeader.Write(dst)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.SystemId[:])
   if err != nil {
      return err
   }
   if b.FullBoxHeader.Version > 0 {
      err := binary.Write(dst, binary.BigEndian, b.KidCount)
      if err != nil {
         return err
      }
      err = binary.Write(dst, binary.BigEndian, b.Kid)
      if err != nil {
         return err
      }
   }
   err = binary.Write(dst, binary.BigEndian, b.DataSize)
   if err != nil {
      return err
   }
   _, err = dst.Write(b.Data)
   if err != nil {
      return err
   }
   return nil
}

// dashif.org/identifiers/content_protection
func (b *Box) Widevine() bool {
   return b.SystemId.String() == "edef8ba979d64acea3c827dcd51d21ed"
}
