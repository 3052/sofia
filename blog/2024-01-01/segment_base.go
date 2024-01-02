package stream

import (
   "bytes"
   "encoding/hex"
   "os"
)

func segment_base() error {
   key, err := hex.DecodeString("dee726e9015a608a3db559a6b9a9c034")
   if err != nil {
      return err
   }
   file, err := os.Create("dec.mp4")
   if err != nil {
      return err
   }
   defer file.Close()
   var (
      sidx uint32 = 1530
      moof uint32 = 16178
   )
   data, err := os.ReadFile("enc.mp4")
   if err != nil {
      return err
   }
   if err := encode_init(file, bytes.NewReader(data[:sidx])); err != nil {
      return err
   }
   file.Write(data[sidx:moof])
   byte_ranges, err := decode_sidx(data, sidx, moof)
   if err != nil {
      return err
   }
   for _, r := range byte_ranges {
      segment := data[r[0]:r[1]+1]
      err := encode_segment(file, bytes.NewReader(segment), key)
      if err != nil {
         return err
      }
   }
   return nil
}
