package main

import (
   "154.pages.dev/sofia/container"
   "os"
)

const name = "../../testdata/youtube/segment-1.0001.m4s"

func main() {
   buf, err := os.ReadFile(name)
   if err != nil {
      panic(err)
   }
   var file container.File
   err = file.Read(buf)
   if err != nil {
      panic(err)
   }
   buf, err = file.Append(nil)
   if err != nil {
      panic(err)
   }
   err = os.WriteFile("segment.m4s", buf, os.ModePerm)
   if err != nil {
      panic(err)
   }
}
