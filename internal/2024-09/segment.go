package main

import (
   "154.pages.dev/sofia/container"
   "os"
)

func main() {
   buf, err := os.ReadFile("segment-1.0001.m4s")
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
   err = os.WriteFile("out.m4s", buf, os.ModePerm)
   if err != nil {
      panic(err)
   }
}
