package sofia

import (
   "os"
   "testing"
)

func Test_Moof(t *testing.T) {
   video, err := os.Open("testdata/amc-audio/segment0.m4f")
   if err != nil {
      t.Fatal(err)
   }
   defer video.Close()
   var f File
   if err := f.Decode(video); err != nil {
      t.Fatal(err)
   }
}
