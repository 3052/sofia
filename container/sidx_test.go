package container

import (
   "fmt"
   "os"
   "testing"
)

func TestSidx(t *testing.T) {
   data, err := os.ReadFile("../testdata/hulu-avc1/init.mp4")
   if err != nil {
      t.Fatal(err)
   }
   var value File
   err = value.Read(data)
   if err != nil {
      t.Fatal(err)
   }
   for _, reference := range value.Sidx.Reference {
      fmt.Println(reference.Size())
   }
}
