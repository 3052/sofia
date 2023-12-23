package sofia

import (
   "encoding/hex"
   "fmt"
   "os"
   "testing"
)

func Test_Mdat(t *testing.T) {
   key, err := hex.DecodeString("c58d3308ed18d43776a78232f552dbe0")
   if err != nil {
      t.Fatal(err)
   }
   media, err := os.Open("testdata/amc-video/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Decode(media); err != nil {
      t.Fatal(err)
   }
   for _, sample := range f.Mdat.Data {
      fmt.Println(sample, key)
   }
}
