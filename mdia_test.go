package sofia

import (
   "os"
   "testing"
)

func Test_Mdia(t *testing.T) {
   src, err := os.Open("testdata/amc-video/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var f File
   if err := f.Decode(src); err != nil {
      t.Fatal(err)
   }
}
