package sofia

import (
   "os"
   "testing"
)

func Test_Moov(t *testing.T) {
   src, err := os.Open("testdata/amc-video/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   dst, err := os.Create("init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer dst.Close()
   var f File
   if err := f.Decode(src); err != nil {
      t.Fatal(err)
   }
   if err := f.Encode(dst); err != nil {
      t.Fatal(err)
   }
}
