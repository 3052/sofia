package sofia

import (
   "fmt"
   "os"
   "testing"
)

func TestByteRanges(t *testing.T) {
   media, err := os.Open("testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer media.Close()
   var value File
   err = value.Read(media)
   if err != nil {
      t.Fatal(err)
   }
   for _, byte_range := range value.SegmentIndex.Ranges(0) {
      fmt.Println(byte_range)
   }
}
