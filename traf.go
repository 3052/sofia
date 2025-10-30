package sofia

import "errors"

type TrafChild struct {
   Tfhd *TfhdBox
   Trun *TrunBox
   Senc *SencBox
   Raw  []byte
}
type TrafBox struct {
   Header   BoxHeader
   RawData  []byte
   Children []TrafChild
}

// Totals calculates the total byte size and duration of all samples in a traf.
func (b *TrafBox) Totals() (totalBytes uint64, totalDuration uint64, err error) {
   trun := b.Trun()
   tfhd := b.Tfhd()
   if trun == nil {
      return 0, 0, errors.New("traf is missing trun box to calculate totals")
   }

   for _, sample := range trun.Samples {
      size := sample.Size
      if size == 0 && tfhd != nil {
         size = tfhd.DefaultSampleSize
      }
      totalBytes += uint64(size)

      duration := sample.Duration
      if duration == 0 && tfhd != nil {
         duration = tfhd.DefaultSampleDuration
      }
      totalDuration += uint64(duration)
   }
   return totalBytes, totalDuration, nil
}

func (b *TrafBox) Parse(data []byte) error {
   if err := b.Header.Parse(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if err := h.Parse(boxData[offset:]); err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return errors.New("invalid child box size in traf")
      }
      childData := boxData[offset : offset+boxSize]
      var child TrafChild
      switch string(h.Type[:]) {
      case "tfhd":
         var tfhd TfhdBox
         if err := tfhd.Parse(childData); err != nil {
            return err
         }
         child.Tfhd = &tfhd
      case "trun":
         var trun TrunBox
         if err := trun.Parse(childData); err != nil {
            return err
         }
         child.Trun = &trun
      case "senc":
         var senc SencBox
         if err := senc.Parse(childData); err != nil {
            return err
         }
         child.Senc = &senc
      default:
         child.Raw = childData
      }
      b.Children = append(b.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return nil
}

func (b *TrafBox) Encode() []byte {
   var content []byte
   for _, child := range b.Children {
      if child.Tfhd != nil {
         content = append(content, child.Tfhd.Encode()...)
      } else if child.Trun != nil {
         content = append(content, child.Trun.Encode()...)
      } else if child.Senc != nil {
         content = append(content, child.Senc.Encode()...)
      } else if child.Raw != nil {
         content = append(content, child.Raw...)
      }
   }
   b.Header.Size = uint32(8 + len(content))
   headerBytes := b.Header.Encode()
   return append(headerBytes, content...)
}

func (b *TrafBox) Tfhd() *TfhdBox {
   for _, child := range b.Children {
      if child.Tfhd != nil {
         return child.Tfhd
      }
   }
   return nil
}

func (b *TrafBox) Trun() *TrunBox {
   for _, child := range b.Children {
      if child.Trun != nil {
         return child.Trun
      }
   }
   return nil
}

func (b *TrafBox) Senc() (*SencBox, bool) {
   for _, child := range b.Children {
      if child.Senc != nil {
         return child.Senc, true
      }
   }
   return nil, false
}
