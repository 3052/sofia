package main

import (
   "154.pages.dev/sofia/container"
   "os"
   "path/filepath"
)
"D:\git\sofia\internal\mp4split\segment-1.0001.m4s"
func read_file(dst []byte, name string) ([]byte, error) {
   src, err := os.ReadFile(name)
   if err != nil {
      return nil, err
   }
   // return append(dst, src...), nil
   var file container.File
   err = file.Read(src)
   if err != nil {
      return nil, err
   }
   return file.Append(dst)
}

func main() {
   dst, err := read_file(nil, "../mp4split/init.mp4")
   if err != nil {
      panic(err)
   }
   matches, err := filepath.Glob("../mp4split/segment-1.*.m4s")
   if err != nil {
      panic(err)
   }
   for _, match := range matches {
      dst, err = read_file(dst, match)
      if err != nil {
         panic(err)
      }
      break
   }
   err = os.WriteFile("out.mp4", dst, os.ModePerm)
   if err != nil {
      panic(err)
   }
}
