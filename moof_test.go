package sofia

import (
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   media, err := os.Open("testdata/amc-audio/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Decode(media); err != nil {
      t.Fatal(err)
   }
}
