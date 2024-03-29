package sofia

import (
   "fmt"
   "os"
   "testing"
)

func TestMovie(t *testing.T) {
   src, err := os.Open("testdata/amc-avc1/init.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var dst File
   if err := dst.Read(src); err != nil {
      t.Fatal(err)
   }
   fmt.Println(dst.Movie.Widevine())
}
