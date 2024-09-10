package file

import (
   "fmt"
   "os"
   "testing"
)

func TestSidx(t *testing.T) {
   in, err := os.Open("testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer in.Close()
   var out File
   err = out.Read(in)
   if err != nil {
      t.Fatal(err)
   }
   for _, data := range out.SegmentIndex.Reference {
      fmt.Println(data.ReferencedSize())
   }
}
