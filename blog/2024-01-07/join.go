package main

import (
   "github.com/Eyevinn/mp4ff/mp4"
   "os"
)

func main() {
   src, err := os.Open("out.mp4")
   if err != nil {
      panic(err)
   }
   defer src.Close()
   dst, err := os.Create("break.mp4")
   if err != nil {
      panic(err)
   }
   defer dst.Close()
   inMp4, err := mp4.DecodeFile(src)
   if err != nil {
      panic(err)
   }
   inMp4.Moov.Trak.Mdia.Minf.Stbl.Stco.ChunkOffset = []uint32{
      40,
   }
   inMp4.Moov.Trak.Mdia.Minf.Stbl.Stsc.Entries = []mp4.StscEntry{
      {
         FirstChunk  : 1,
         SamplesPerChunk : 14416,
      },
   }
   if err := inMp4.Encode(dst); err != nil {
      panic(err)
   }
}
