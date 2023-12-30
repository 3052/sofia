package sofia

import (
   "fmt"
   "os"
   "testing"
)

func Test_Sidx(t *testing.T) {
   media, err := os.Open("testdata/hulu-video/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var f File
   if err := f.Decode(media); err != nil {
      t.Fatal(err)
   }
   for _, byte_range := range f.SegmentIndex.ByteRanges(0) {
      fmt.Println(byte_range)
   }
}
