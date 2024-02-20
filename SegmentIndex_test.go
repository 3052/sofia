package sofia

import (
   "fmt"
   "os"
   "testing"
)

func TestGlobal(t *testing.T) {
   for i := 1; i <= 915; i++ {
      fmt.Printf("seg_%v.m4s\n", i)
   }
}

func TestByteRanges(t *testing.T) {
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
