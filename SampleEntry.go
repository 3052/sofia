package sofia

import (
   "encoding/binary"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) abstract class SampleEntry(
//      unsigned int(32) format
//   ) extends Box(format) {
//      const unsigned int(8)[6] reserved = 0;
//      unsigned int(16) data_reference_index;
//   }
type SampleEntry struct {
   BoxHeader          BoxHeader
   Reserved           [6]uint8
   DataReferenceIndex uint16
}

func (s *SampleEntry) Read(r io.Reader) error {
   _, err := io.ReadFull(r, s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Read(r, binary.BigEndian, &s.DataReferenceIndex)
}

func (s *SampleEntry) Write(w io.Writer) error {
   err := s.BoxHeader.Write(w)
   if err != nil {
      return err
   }
   _, err = w.Write(s.Reserved[:])
   if err != nil {
      return err
   }
   return binary.Write(w, binary.BigEndian, s.DataReferenceIndex)
}
