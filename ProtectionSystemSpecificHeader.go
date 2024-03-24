package sofia

import (
   "encoding/binary"
   "io"
)

// ISO/IEC 23001-7
//  aligned(8) class ProtectionSystemSpecificHeaderBox extends FullBox(
//     'pssh', version, flags=0,
//  ) {
//     unsigned int(8)[16] SystemID;
//     if (version > 0) {
//        unsigned int(32) KID_count;
//        {
//           unsigned int(8)[16] KID;
//        } [KID_count];
//     }
//     unsigned int(32) DataSize;
//     unsigned int(8)[DataSize] Data;
//  }
type ProtectionSystemSpecificHeader struct {
   BoxHeader  BoxHeader
   FullBoxHeader FullBoxHeader
   SystemId UUID
   DataSize uint32
   Data []uint8
}

func (p *ProtectionSystemSpecificHeader) read(r io.Reader) error {
   err := p.FullBoxHeader.read(r)
   if err != nil {
      return err
   }
   if _, err := io.ReadFull(r, p.SystemId[:]); err != nil {
      return err
   }
   if err := binary.Read(r, binary.BigEndian, &p.DataSize); err != nil {
      return err
   }
   p.Data = make([]uint8, p.DataSize)
   if _, err := io.ReadFull(r, p.Data); err != nil {
      return err
   }
   return nil
}

func (p ProtectionSystemSpecificHeader) write(w io.Writer) error {
   err := p.BoxHeader.write(w)
   if err != nil {
      return err
   }
   if err := p.FullBoxHeader.write(w); err != nil {
      return err
   }
   if _, err := w.Write(p.SystemId[:]); err != nil {
      return err
   }
   if err := binary.Write(w, binary.BigEndian, p.DataSize); err != nil {
      return err
   }
   if _, err := w.Write(p.Data); err != nil {
      return err
   }
   return nil
}
