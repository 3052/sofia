package file

import (
   "fmt"
   "os"
   "testing"
)

func TestSidx(t *testing.T) {
   src, err := os.Open("../testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   defer src.Close()
   var value File
   err = value.Read(src)
   if err != nil {
      t.Fatal(err)
   }
   for _, reference := range value.SegmentIndex.Reference {
      fmt.Println(reference.ReferencedSize())
   }
}
