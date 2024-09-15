package main

import (
   "154.pages.dev/sofia/container"
   "os"
)

func main() {
   buf, err := os.ReadFile("../mp4split/init.mp4")
   if err != nil {
      panic(err)
   }
   var file container.File
   err = file.Read(buf)
   if err != nil {
      panic(err)
   }
}
