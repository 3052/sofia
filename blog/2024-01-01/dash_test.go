package stream

import (
   "154.pages.dev/sofia"
   "os"
   "testing"
)

func Test_Copy(t *testing.T) {
   err := file_copy()
   if err != nil {
      t.Fatal(err)
   }
}

func Test_Decrypt(t *testing.T) {
   err := segment_base()
   if err != nil {
      t.Fatal(err)
   }
}

func Test_Segment(t *testing.T) {
   src, err := os.Open("segment-1.0001.m4s")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   dst, err := os.Create("out.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   var file sofia.File
   if err := file.Decode(src); err != nil {
      t.Fatal(err)
   }
   if err := file.Encode(dst); err != nil {
      t.Fatal(err)
   }
}

func Test_Init(t *testing.T) {
   src, err := os.Open("init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   dst, err := os.Create("out.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   var file sofia.File
   if err := file.Decode(src); err != nil {
      t.Fatal(err)
   }
   if err := file.Encode(dst); err != nil {
      t.Fatal(err)
   }
}

