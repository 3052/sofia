package sofia

import (
   "encoding/binary"
   "errors"
   "io"
)

// ISO/IEC 14496-12
//   aligned(8) class SampleDescriptionBox() extends FullBox('stsd', version, 0) {
//      int i ;
//      unsigned int(32) entry_count;
//      for (i = 1 ; i <= entry_count ; i++){
//         SampleEntry(); // an instance of a class derived from SampleEntry
//      }
//   }
type SampleDescription struct {
   BoxHeader     BoxHeader
   FullBoxHeader FullBoxHeader
   EntryCount    uint32
   Boxes         []Box
   AudioSample   *AudioSampleEntry
   VisualSample  *VisualSampleEntry
}

func (s SampleDescription) Protection() (*ProtectionSchemeInfo, bool) {
   if v := s.AudioSample; v != nil {
      return &v.ProtectionScheme, true
   }
   if v := s.VisualSample; v != nil {
      return &v.ProtectionScheme, true
   }
   return nil, false
}

func (s SampleDescription) SampleEntry() (*SampleEntry, bool) {
   if v := s.AudioSample; v != nil {
      return &v.SampleEntry, true
   }
   if v := s.VisualSample; v != nil {
      return &v.SampleEntry, true
   }
   return nil, false
}

func (s *SampleDescription) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   err := s.FullBoxHeader.read(r)
   if err != nil {
      return err
   }
   err = binary.Read(r, binary.BigEndian, &s.EntryCount)
   if err != nil {
      return err
   }
   for {
      var head BoxHeader
      err := head.Read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      _, size := head.get_size()
      switch head.debug() {
      case "avc1", // Tubi
      "ec-3", // Max
      "mp4a": // Tubi
         object := Box{BoxHeader: head}
         err := object.read(r)
         if err != nil {
            return err
         }
         s.Boxes = append(s.Boxes, object)
      case "enca":
         s.AudioSample = new(AudioSampleEntry)
         s.AudioSample.SampleEntry.BoxHeader = head
         err := s.AudioSample.read(r, size)
         if err != nil {
            return err
         }
      case "encv":
         s.VisualSample = new(VisualSampleEntry)
         s.VisualSample.SampleEntry.BoxHeader = head
         err := s.VisualSample.read(r, size)
         if err != nil {
            return err
         }
      default:
         return errors.New("SampleDescription.read")
      }
   }
}

func (s SampleDescription) write(w io.Writer) error {
   err := s.BoxHeader.write(w)
   if err != nil {
      return err
   }
   err = s.FullBoxHeader.write(w)
   if err != nil {
      return err
   }
   err = binary.Write(w, binary.BigEndian, s.EntryCount)
   if err != nil {
      return err
   }
   for _, object := range s.Boxes {
      err := object.write(w)
      if err != nil {
         return err
      }
   }
   if s.AudioSample != nil {
      err := s.AudioSample.write(w)
      if err != nil {
         return err
      }
   }
   if s.VisualSample != nil {
      err := s.VisualSample.write(w)
      if err != nil {
         return err
      }
   }
   return nil
}
