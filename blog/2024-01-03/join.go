package main

import (
   "io"
   "os"
)

func write(w io.Writer, name string) error {
   b, err := os.ReadFile(name)
   if err != nil {
      return err
   }
   if _, err := w.Write(b); err != nil {
      return err
   }
   return nil
}

var parts = []string{
   "init.mp4",
   "segment-1.0001.m4s",
   "segment-1.0002.m4s",
}

func main() {
   file, err := os.Create("frag.mp4")
   if err != nil {
      panic(err)
   }
   defer file.Close()
   for _, part := range parts {
      err := write(file, part)
      if err != nil {
         panic(err)
      }
   }
}
