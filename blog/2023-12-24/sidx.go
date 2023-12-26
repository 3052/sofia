package sidx

import (
   "github.com/yapingcat/gomedia/go-mp4"
   "io"
)

func byte_ranges(r io.Reader, start uint32) ([][]uint32, error) {
   sidx := mp4.SegmentIndexBox{
      Box: &mp4.FullBox{
         Box: &mp4.BasicBox{},
      },
   }
   if _, err := sidx.Box.Box.Decode(r); err != nil {
      return nil, err
   }
   if _, err := sidx.Decode(r); err != nil {
      return nil, err
   }
   var rs [][]uint32
   for _, e := range sidx.Entrys {
      r := []uint32{start, start + e.ReferencedSize - 1}
      rs = append(rs, r)
      start += e.ReferencedSize
   }
   return rs, nil
}
