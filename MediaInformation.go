package sofia

import (
   "errors"
   "io"
)

// ISO/IEC 14496-12
//
//   aligned(8) class MediaInformationBox extends Box('minf') {
//   }
type MediaInformation struct {
   BoxHeader   BoxHeader
   Boxes       []Box
   SampleTable SampleTable
}

func (m *MediaInformation) read(r io.Reader, size int64) error {
   r = io.LimitReader(r, size)
   for {
      var head BoxHeader
      err := head.read(r)
      if err == io.EOF {
         return nil
      } else if err != nil {
         return err
      }
      switch head.debug() {
      case "stbl":
         _, size := head.get_size()
         m.SampleTable.BoxHeader = head
         err := m.SampleTable.read(r, size)
         if err != nil {
            return err
         }
      case "dinf", // Roku
         "smhd", // Roku
         "vmhd": // Roku
         b := Box{BoxHeader: head}
         err := b.read(r)
         if err != nil {
            return err
         }
         m.Boxes = append(m.Boxes, b)
      default:
         return errors.New("MediaInformation.read")
      }
   }
}

func (m MediaInformation) write(w io.Writer) error {
   err := m.BoxHeader.write(w)
   if err != nil {
      return err
   }
   for _, b := range m.Boxes {
      err := b.write(w)
      if err != nil {
         return err
      }
   }
   return m.SampleTable.write(w)
}
