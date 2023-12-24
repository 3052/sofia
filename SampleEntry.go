package sofia

import (
   "encoding/binary"
   "fmt"
   "io"
)

// All SampleEntry boxes may contain “extra boxes” not explicitly defined in the
// box syntax of this or derived specifications. When present, such boxes shall
// follow all defined fields and should follow any defined contained boxes.
// Decoders shall presume a sample entry box could contain extra boxes and shall
// continue parsing as though they are present until the containing box length is
// exhausted.
//
// aligned(8) abstract class SampleEntry(unsigned int(32) format) extends Box(format) {
//    const unsigned int(8)[6] reserved = 0;
//    unsigned int(16) data_reference_index;
// }
type SampleEntry struct {
   Header BoxHeader
   Reserved [6]uint8
   Data_Reference_Index uint16
   Boxes []*Box
}

func (s *SampleEntry) Decode(r io.Reader) error {
   _, err := r.Read(s.Reserved[:])
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.Data_Reference_Index)
   if err != nil {
      return err
   }
   for {
      var head BoxHeader
      err := head.Decode(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      size := head.BoxPayload()
      switch head.Type() {
      case "sinf":
         value := Box{Header: head}
         value.Payload = make([]byte, size)
         _, err := r.Read(value.Payload)
         if err != nil {
            return err
         }
         s.Boxes = append(s.Boxes, &value)
      default:
         return fmt.Errorf("%q", head.RawType)
      }
   }
}

func (s SampleEntry) Encode(w io.Writer) error {
   err := s.Header.Encode(w)
   if err != nil {
      return err
   }
   if _, err := w.Write(s.Reserved[:]); err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.Data_Reference_Index)
   if err != nil {
      return err
   }
   for _, value := range s.Boxes {
      err := value.Encode(w)
      if err != nil {
         return err
      }
   }
   return nil
}
