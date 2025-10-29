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

// getTotals is an unexported helper to calculate the total byte size and
// duration of all samples in a traf. It avoids looping through the samples
// multiple times.
func (b *TrafBox) getTotals() (totalBytes uint64, totalDuration uint64, err error) {
   trun := b.GetTrun()
   tfhd := b.GetTfhd()
   if trun == nil {
      return 0, 0, errors.New("traf is missing trun box to calculate totals")
   }

   for _, sample := range trun.Samples {
      // Calculate size for bandwidth
      size := sample.Size
      if size == 0 && tfhd != nil {
         size = tfhd.DefaultSampleSize
      }
      totalBytes += uint64(size)

      // Calculate duration
      duration := sample.Duration
      if duration == 0 && tfhd != nil {
         duration = tfhd.DefaultSampleDuration
      }
      totalDuration += uint64(duration)
   }
   return totalBytes, totalDuration, nil
}

// GetTotalDuration calculates the total duration of all samples in the traf.
func (b *TrafBox) GetTotalDuration() (uint64, error) {
   _, totalDuration, err := b.getTotals()
   return totalDuration, err
}

// GetBandwidth calculates the average bandwidth of the traf in bits per second.
func (b *TrafBox) GetBandwidth(timescale uint32) (uint64, error) {
   if timescale == 0 {
      return 0, errors.New("timescale cannot be zero")
   }

   totalBytes, totalDuration, err := b.getTotals()
   if err != nil {
      return 0, err
   }

   if totalDuration == 0 {
      // Avoid division by zero if the duration is unknown.
      return 0, nil
   }

   // Bandwidth in bps = (TotalBytes * 8 bits/byte) / (TotalDuration / Timescale in seconds)
   // Simplified: (TotalBytes * 8 * Timescale) / TotalDuration
   bandwidth := (totalBytes * 8 * uint64(timescale)) / totalDuration
   return bandwidth, nil
}

// Parse parses the 'traf' box from a byte slice.
func (b *TrafBox) Parse(data []byte) error {
   if _, err := b.Header.Read(data); err != nil {
      return err
   }
   b.RawData = data[:b.Header.Size]
   boxData := data[8:b.Header.Size]
   offset := 0
   for offset < len(boxData) {
      var h BoxHeader
      if _, err := h.Read(boxData[offset:]); err != nil {
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
   encoded := make([]byte, b.Header.Size)
   b.Header.Write(encoded)
   copy(encoded[8:], content)
   return encoded
}
func (b *TrafBox) GetTfhd() *TfhdBox {
   for _, child := range b.Children {
      if child.Tfhd != nil {
         return child.Tfhd
      }
   }
   return nil
}
func (b *TrafBox) GetTrun() *TrunBox {
   for _, child := range b.Children {
      if child.Trun != nil {
         return child.Trun
      }
   }
   return nil
}
func (b *TrafBox) GetSenc() *SencBox {
   for _, child := range b.Children {
      if child.Senc != nil {
         return child.Senc
      }
   }
   return nil
}
