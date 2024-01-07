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
   if err := inMp4.Encode(dst); err != nil {
      panic(err)
   }
}
