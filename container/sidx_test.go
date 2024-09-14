package container

import (
   "fmt"
   "os"
   "testing"
)

func TestSidx(t *testing.T) {
   buf, err := os.ReadFile("../testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   var value File
   err = value.Decode(buf)
   if err != nil {
      t.Fatal(err)
   }
   for _, reference := range value.Sidx.Reference {
      fmt.Println(reference.ReferencedSize())
   }
}
