package mp4

import (
   "errors"
   "fmt"
)

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

// GetBandwidth calculates the bandwidth of the track fragment in bits per second.
// It requires the track's timescale from the 'mdhd' box.
func (b *TrafBox) GetBandwidth(timescale uint32) (uint64, error) {
   trun := b.GetTrun()
   tfhd := b.GetTfhd()

   if trun == nil {
      return 0, errors.New("traf is missing trun box")
   }
   if timescale == 0 {
      return 0, errors.New("timescale cannot be zero")
   }

   var totalBytes uint64 = 0
   var totalDuration uint64 = 0

   for _, sample := range trun.Samples {
      size := sample.Size
      // If per-sample size is not present, use the default from tfhd
      if size == 0 && tfhd != nil {
         size = tfhd.DefaultSampleSize
      }
      totalBytes += uint64(size)

      duration := sample.Duration
      // If per-sample duration is not present, use the default from tfhd
      if duration == 0 && tfhd != nil {
         duration = tfhd.DefaultSampleDuration
      }
      totalDuration += uint64(duration)
   }

   if totalDuration == 0 {
      return 0, nil // Avoid division by zero if fragment has no duration
   }

   // Bandwidth in bits per second:
   // (totalBytes * 8 bits/byte) / (totalDuration / timescale seconds)
   // which simplifies to: (totalBytes * 8 * timescale) / totalDuration
   bandwidth := (totalBytes * 8 * uint64(timescale)) / totalDuration
   return bandwidth, nil
}

// ParseTraf and other helpers remain the same...
func ParseTraf(data []byte) (TrafBox, error) {
   header, _, err := ReadBoxHeader(data)
   if err != nil {
      return TrafBox{}, err
   }
   var traf TrafBox
   traf.Header = header
   traf.RawData = data[:header.Size]
   boxData := data[8:header.Size]
   offset := 0
   for offset < len(boxData) {
      h, _, err := ReadBoxHeader(boxData[offset:])
      if err != nil {
         break
      }
      boxSize := int(h.Size)
      if boxSize == 0 {
         boxSize = len(boxData) - offset
      }
      if boxSize < 8 || offset+boxSize > len(boxData) {
         return TrafBox{}, fmt.Errorf("invalid child box size in traf")
      }
      childData := boxData[offset : offset+boxSize]
      var child TrafChild
      switch string(h.Type[:]) {
      case "tfhd":
         tfhd, err := ParseTfhd(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Tfhd = &tfhd
      case "trun":
         trun, err := ParseTrun(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Trun = &trun
      case "senc":
         senc, err := ParseSenc(childData)
         if err != nil {
            return TrafBox{}, err
         }
         child.Senc = &senc
      default:
         child.Raw = childData
      }
      traf.Children = append(traf.Children, child)
      offset += boxSize
      if h.Size == 0 {
         break
      }
   }
   return traf, nil
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
