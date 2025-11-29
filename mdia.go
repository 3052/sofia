package sofia

import "encoding/binary"

type MdiaChild struct {
   Mdhd *MdhdBox
   Minf *MinfBox
   Raw  []byte
}

type MdiaBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []MdiaChild
}

func (b *MdiaBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   return parseContainer(data[8:b.Header.Size], func(h BoxHeader, content []byte) error {
      var child MdiaChild
      switch string(h.Type[:]) {
      case "mdhd":
         var mdhd MdhdBox
         if err := mdhd.Parse(content); err != nil {
            return err
         }
         child.Mdhd = &mdhd
      case "minf":
         var minf MinfBox
         if err := minf.Parse(content); err != nil {
            return err
         }
         child.Minf = &minf
      default:
         child.Raw = content
      }
      b.Children = append(b.Children, child)
      return nil
   })
}

func (b *MdiaBox) Encode() []byte {
   buf := make([]byte, 8)
   for _, child := range b.Children {
      if child.Mdhd != nil {
         buf = append(buf, child.Mdhd.RawData...)
      } else if child.Minf != nil {
         buf = append(buf, child.Minf.Encode()...)
      } else if child.Raw != nil {
         buf = append(buf, child.Raw...)
      }
   }
   b.Header.Size = uint32(len(buf))
   binary.BigEndian.PutUint32(buf[0:4], b.Header.Size)
   copy(buf[4:8], b.Header.Type[:])
   return buf
}

func (b *MdiaBox) Mdhd() (*MdhdBox, bool) {
   for _, child := range b.Children {
      if child.Mdhd != nil {
         return child.Mdhd, true
      }
   }
   return nil, false
}

func (b *MdiaBox) Minf() (*MinfBox, bool) {
   for _, child := range b.Children {
      if child.Minf != nil {
         return child.Minf, true
      }
   }
   return nil, false
}
