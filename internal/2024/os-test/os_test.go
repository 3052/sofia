package main

import (
   "io"
   "os"
   "testing"
)

func BenchmarkTruncate(b *testing.B) {
   for range b.N {
      file, err := os.Create("hello.txt")
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.WriteString("alfa")
      if err != nil {
         b.Fatal(err)
      }
      err = file.Truncate(
         int64(len("alfa bravo")),
      )
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.Seek(0, io.SeekEnd)
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.WriteString(" charlie\n")
      if err != nil {
         b.Fatal(err)
      }
      err = file.Close()
      if err != nil {
         b.Fatal(err)
      }
   }
}

func BenchmarkWrite(b *testing.B) {
   for range b.N {
      file, err := os.Create("hello.txt")
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.WriteString("alfa")
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.WriteString(" bravo")
      if err != nil {
         b.Fatal(err)
      }
      _, err = file.WriteString(" charlie\n")
      if err != nil {
         b.Fatal(err)
      }
      err = file.Close()
      if err != nil {
         b.Fatal(err)
      }
   }
}
