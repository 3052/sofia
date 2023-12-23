package sofia

import (
   "fmt"
   "os"
   "testing"
)

func Test_Mdat(t *testing.T) {
   media, err := os.Open("testdata/amc-video/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Decode(media); err != nil {
      t.Fatal(err)
   }
   for _, data := range f.Mdat.Data {
      fmt.Println(len(data))
   }
   fmt.Println("len(f.Mdat.Data)", len(f.Mdat.Data))
}
